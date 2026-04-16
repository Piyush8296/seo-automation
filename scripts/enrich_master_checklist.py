import re
import sys
import zipfile
from copy import deepcopy
from pathlib import Path
import xml.etree.ElementTree as ET


NS_MAIN = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
NS_REL = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
NS_PKG_REL = "http://schemas.openxmlformats.org/package/2006/relationships"
NS = {"a": NS_MAIN, "r": NS_REL, "pr": NS_PKG_REL}

ET.register_namespace("", NS_MAIN)
ET.register_namespace("r", NS_REL)


def col_to_idx(col: str) -> int:
    n = 0
    for ch in col:
        if "A" <= ch <= "Z":
            n = n * 26 + (ord(ch) - 64)
    return n - 1


def idx_to_col(idx: int) -> str:
    idx += 1
    out = []
    while idx:
        idx, rem = divmod(idx - 1, 26)
        out.append(chr(65 + rem))
    return "".join(reversed(out))


def parse_cell_value(cell: ET.Element, shared_strings: list[str]) -> str:
    cell_type = cell.attrib.get("t")
    value = cell.find("a:v", NS)
    if cell_type == "s" and value is not None and value.text is not None:
        return shared_strings[int(value.text)]
    if cell_type == "inlineStr":
        txt = cell.find("a:is/a:t", NS)
        return txt.text if txt is not None and txt.text is not None else ""
    if value is not None and value.text is not None:
        return value.text
    return ""


def read_shared_strings(zf: zipfile.ZipFile) -> list[str]:
    shared = []
    if "xl/sharedStrings.xml" not in zf.namelist():
        return shared
    root = ET.fromstring(zf.read("xl/sharedStrings.xml"))
    for si in root.findall("a:si", NS):
        shared.append("".join((t.text or "") for t in si.iterfind(".//a:t", NS)))
    return shared


def workbook_sheet_targets(zf: zipfile.ZipFile) -> dict[str, str]:
    rels_root = ET.fromstring(zf.read("xl/_rels/workbook.xml.rels"))
    rels = {
        rel.attrib["Id"]: rel.attrib["Target"]
        for rel in rels_root.findall("pr:Relationship", NS)
    }

    wb_root = ET.fromstring(zf.read("xl/workbook.xml"))
    targets = {}
    for sheet in wb_root.find("a:sheets", NS):
        name = sheet.attrib["name"]
        rid = sheet.attrib[f"{{{NS_REL}}}id"]
        target = rels[rid]
        if not target.startswith("xl/"):
            target = "xl/" + target
        targets[name] = target
    return targets


def sheet_rows(sheet_root: ET.Element, shared_strings: list[str]) -> list[list[str]]:
    rows = []
    for row in sheet_root.findall(".//a:sheetData/a:row", NS):
        vals = {}
        max_idx = -1
        for cell in row.findall("a:c", NS):
            ref = cell.attrib.get("r", "")
            match = re.match(r"([A-Z]+)", ref)
            idx = col_to_idx(match.group(1)) if match else len(vals)
            max_idx = max(max_idx, idx)
            vals[idx] = parse_cell_value(cell, shared_strings)
        rows.append([vals.get(i, "") for i in range(max_idx + 1)])
    return rows


def build_inline_cell(ref: str, text: str) -> ET.Element:
    cell = ET.Element(f"{{{NS_MAIN}}}c", {"r": ref, "t": "inlineStr"})
    is_elem = ET.SubElement(cell, f"{{{NS_MAIN}}}is")
    t_elem = ET.SubElement(is_elem, f"{{{NS_MAIN}}}t")
    t_elem.text = text
    return cell


def parse_check_id(check_id: str) -> tuple[str, int]:
    match = re.match(r"([A-Z]+)-(\d+)$", check_id or "")
    if not match:
        return "", 0
    return match.group(1), int(match.group(2))


def classify_row(row: dict[str, str]) -> tuple[str, str]:
    check_id = row["Check ID"]
    name = row["Check Name"]
    check_type = row["Type"]
    prefix, num = parse_check_id(check_id)

    def partial(added: str, missing: str) -> tuple[str, str]:
        return "Partial", f"Added: {added} Missing: {missing}"

    manual_like_prefixes = {"BACKLINK", "BRAND", "ANALYTICS", "TRAFFIC", "LOG"}
    if prefix in manual_like_prefixes:
        return "Manual only", ""

    if prefix == "CANONICAL":
        if num in {1, 6, 17}:
            return "Yes", ""
        if num in {2, 3, 4, 12, 13, 18, 19}:
            return partial(
                "basic canonical, AMP, duplicate-content, and hreflang-adjacent checks exist.",
                "the exact PRD rule is not covered end-to-end."
            )
        return "No", ""

    if prefix == "ROBOTS":
        if num in {1, 4}:
            return partial(
                "crawler-level robots.txt metadata is collected for file presence and sitemap directive discovery.",
                "these signals are not exposed as first-class checklist findings in the current report/check pipeline."
            )
        if num in {5, 12, 13}:
            return partial(
                "robots-aware crawling and crawl-budget/search/faceted checks exist.",
                "the exact robots.txt rule is not directly validated."
            )
        if num == 10:
            return "Manual only", ""
        return "No", ""

    if prefix == "SITEMAP":
        if num in {1, 10, 11}:
            return "Yes", ""
        if num in {3, 7}:
            return partial(
                "the crawler parses sitemap XML and validates some sitemap URL outcomes.",
                "there is no dedicated XML-validity finding, and 5xx/non-crawled sitemap URLs are not fully covered."
            )
        if num == 20:
            return partial(
                "orphan-page and sitemap coverage checks exist.",
                "there is no dedicated sitemap-orphan cross-check."
            )
        if num in {2, 19}:
            return "Manual only", ""
        return "No", ""

    if prefix == "SSL":
        if num in {1, 5, 6}:
            return "Yes", ""
        return "No", ""

    if prefix == "REDIRECT":
        if num in {1, 2, 3}:
            return "Yes", ""
        if num in {4, 8}:
            return partial(
                "generic redirect-chain/loop logic exists.",
                "the exact redirect scenario is not explicitly implemented."
            )
        return "No", ""

    if prefix == "STATUS":
        if num == 1:
            return "No", "Missing: internal link status codes are not currently populated, so the broken-link checks do not run end-to-end."
        if num == 3:
            return partial(
                "5xx responses are detected during crawl.",
                "there is no dedicated monitoring/alert workflow."
            )
        return "No", ""

    if prefix == "URL":
        if num in {1, 2, 3, 14}:
            return "Yes", ""
        if num in {4, 5, 8, 9, 13}:
            return partial(
                "related URL-depth, parameter, case, or faceted checks exist.",
                "the exact rule or threshold differs from the checklist."
            )
        return "No", ""

    if prefix == "CRAWL":
        if num in {1, 8}:
            return "Yes", ""
        if num in {2, 3, 5, 12}:
            return partial(
                "related crawl-budget, pagination, depth, and thin-content checks exist.",
                "the exact crawlability rule is only partially covered."
            )
        if num in {4, 9, 10, 11, 15}:
            return "Manual only", ""
        return "No", ""

    if prefix == "TITLE":
        if num in {1, 2}:
            return "Yes", ""
        if num in {3, 11, 15}:
            return partial(
                "title presence/duplicate/length and title-vs-H1 heuristics exist.",
                "the exact optimization rule is not fully implemented."
            )
        if num == 18:
            return "Manual only", ""
        return "No", ""

    if prefix == "META":
        if num in {1, 2}:
            return "Yes", ""
        if num == 3:
            return partial(
                "meta description length checks exist.",
                "the threshold is broader than the checklist target."
            )
        if num == 7:
            return "Manual only", ""
        return "No", ""

    if prefix == "H1":
        if num in {1, 2, 4, 6, 8}:
            return "Yes", ""
        if num in {5, 7}:
            return partial(
                "title/H1 and hierarchy checks exist.",
                "the exact checklist rule is only indirectly covered."
            )
        return "No", ""

    if prefix == "IMG":
        if num in {1, 15}:
            return "Yes", ""
        if num in {2, 4, 5, 8, 9}:
            return partial(
                "alt-text quality, filename-as-alt, and lazy-load heuristics exist.",
                "the exact checklist requirement is not fully validated."
            )
        if num == 10:
            return "Manual only", ""
        return "No", ""

    if prefix == "INTLINK":
        if num in {3, 9, 14}:
            return "Yes", ""
        if num == 4:
            return "No", "Missing: internal link status codes are not currently fetched into the page model, so 404 internal-link detection is scaffolded but not operational."
        if num in {2, 7, 10}:
            return partial(
                "anchor quality, breadcrumb, and depth-related checks exist.",
                "the checklist asks for a more specific implementation."
            )
        return "No", ""

    if prefix == "SCHEMA":
        if num == 1:
            return "Yes", ""
        if num in {3, 4, 11, 13, 15, 16}:
            return partial(
                "generic JSON-LD and Article/Product/Breadcrumb/FAQ validation exists.",
                "the schema family or validation depth is narrower than the checklist."
            )
        if num in {8, 17}:
            return "Manual only", ""
        return "No", ""

    if prefix == "KW":
        if num == 9:
            return partial(
                "URL, title, H1, and body-related checks exist separately.",
                "there is no unified keyword-alignment implementation."
            )
        if check_type == "Manual":
            return "Manual only", ""
        return "No", ""

    if prefix == "CONTENT":
        if num == 1:
            return partial(
                "thin-content and very-thin-content checks exist.",
                "the checklist asks for key-page minimum word-count validation."
            )
        if num in {2, 3, 4, 7, 8}:
            return partial(
                "near-duplicate, thin-content, and E-E-A-T heuristics exist.",
                "the checklist requires broader content-quality validation."
            )
        return "No", ""

    if prefix == "DUPE":
        if num in {1, 3}:
            return partial(
                "parameter, mobile/desktop, and duplicate-content heuristics exist.",
                "the exact duplicate-content case is not fully implemented."
            )
        if num in {4, 5, 8, 9, 10}:
            return partial(
                "near-duplicate, archive, and pagination-value heuristics exist.",
                "the checklist asks for more targeted page-type coverage."
            )
        if num == 6:
            return partial(
                "low-value and crawl-budget checks exist.",
                "there is no explicit auto-generated-page detector."
            )
        return "No", ""

    if prefix == "CWV":
        if num in {1, 2, 3, 5, 6, 7, 10}:
            return partial(
                "HTML heuristics for LCP, CLS, INP/FID, and response-time risks exist.",
                "there is no real measured CWV/Lighthouse/RUM implementation."
            )
        if num == 4:
            return "Manual only", ""
        return "No", ""

    if prefix == "SPEED":
        if num == 16:
            return "Yes", ""
        if num in {5, 6, 7, 8, 13, 15, 18, 19, 20, 23, 26}:
            return partial(
                "related performance heuristics exist in the crawler.",
                "the checklist asks for more exact measurement or broader coverage."
            )
        if num in {4, 28}:
            return "Manual only", ""
        return "No", ""

    if prefix == "MOBILE":
        if num == 2:
            return "Yes", ""
        if num in {1, 4, 7, 11}:
            return partial(
                "basic mobile, font-size, mobile-vs-desktop, and AMP-related checks exist.",
                "the checklist asks for fuller device or UX validation."
            )
        if num in {6, 12, 13, 14}:
            return "Manual only", ""
        return "No", ""

    if prefix == "LOCAL":
        if num == 7:
            return partial(
                "generic Organization/LocalBusiness schema support exists.",
                "there is no dedicated location-page LocalBusiness validation."
            )
        if num in {1, 2, 3, 4, 5, 6, 9, 11, 12, 14}:
            return "Manual only", ""
        return "No", ""

    if prefix == "DROP":
        if num == 6:
            return "Yes", ""
        if num in {5, 7, 8, 14, 15, 16}:
            return partial(
                "related robots, canonical, performance, redirect, server-error, or hreflang checks exist.",
                "there is no dedicated ranking-drop investigation workflow."
            )
        if num in {1, 2, 3, 4, 10, 11, 12, 13, 17, 18, 20}:
            return "Manual only", ""
        return "No", ""

    if prefix == "TRAFFIC":
        return "Manual only", ""

    if prefix == "INDEX":
        if num in {6, 8}:
            return partial(
                "noindex, low-value-page, and crawl-budget checks exist.",
                "the checklist requires fuller index-coverage analysis."
            )
        return "Manual only", ""

    if prefix == "LOG":
        return "Manual only", ""

    if prefix == "JS":
        if num in {1, 9}:
            return partial(
                "raw-HTML and third-party-script heuristics exist.",
                "there is no browser-rendered JavaScript SEO implementation."
            )
        if num in {4, 10}:
            return "Manual only", ""
        return "No", ""

    if prefix == "HREFLANG":
        if num in {2, 3, 4}:
            return "Yes", ""
        if num == 1:
            return partial(
                "hreflang validation checks exist.",
                "there is no dedicated positive presence validator for all multilingual pages."
            )
        if num == 7:
            return "Manual only", ""
        return "No", ""

    if prefix in {"BACKLINK", "BRAND", "ANALYTICS"}:
        return "Manual only", ""

    if check_type in {"Manual", "Monitoring", "UX"}:
        return "Manual only", ""

    return "No", ""


def enrich_master_checklist(src: Path, dst: Path) -> tuple[int, dict[str, int]]:
    with zipfile.ZipFile(src) as zf:
        targets = workbook_sheet_targets(zf)
        master_path = targets["MASTER CHECKLIST"]
        shared_strings = read_shared_strings(zf)
        sheet_root = ET.fromstring(zf.read(master_path))
        rows = sheet_rows(sheet_root, shared_strings)

        header = rows[0]
        wanted = header + ["Implemented", "Implementation Notes"]

        summary = {"Yes": 0, "Partial": 0, "No": 0, "Manual only": 0}

        sheet_data = sheet_root.find(".//a:sheetData", NS)
        xml_rows = sheet_data.findall("a:row", NS)

        for idx, xml_row in enumerate(xml_rows, start=1):
            row_vals = rows[idx - 1]
            row_dict = {wanted[i]: row_vals[i] if i < len(row_vals) else "" for i in range(len(header))}
            if idx == 1:
                implemented, note = "Implemented", "Implementation Notes"
            else:
                implemented, note = classify_row(row_dict)
                summary[implemented] += 1

            xml_row.append(build_inline_cell(f"K{idx}", implemented))
            xml_row.append(build_inline_cell(f"L{idx}", note))

        dimension = sheet_root.find("a:dimension", NS)
        if dimension is not None:
            max_row = len(xml_rows)
            dimension.attrib["ref"] = f"A1:L{max_row}"

        data = ET.tostring(sheet_root, encoding="utf-8", xml_declaration=True)

        with zipfile.ZipFile(dst, "w", zipfile.ZIP_DEFLATED) as out:
            for info in zf.infolist():
                payload = data if info.filename == master_path else zf.read(info.filename)
                out.writestr(info, payload)

    return len(xml_rows) - 1, summary


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: python3 scripts/enrich_master_checklist.py <src.xlsx> <dst.xlsx>")
        return 2

    src = Path(sys.argv[1])
    dst = Path(sys.argv[2])
    row_count, summary = enrich_master_checklist(src, dst)
    print(f"enriched rows: {row_count}")
    for key in ["Yes", "Partial", "No", "Manual only"]:
        print(f"{key}: {summary[key]}")
    print(f"output: {dst}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
