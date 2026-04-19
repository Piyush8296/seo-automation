#!/bin/bash
set -e

# Ensure Cloud SDK binaries are discoverable in Cloud Run.
export PATH="/opt/google-cloud-sdk/bin:/root/google-cloud-sdk/bin:${PATH}"
GSUTIL_BIN="${GSUTIL_BIN:-$(command -v gsutil 2>/dev/null || true)}"
if [[ -z "$GSUTIL_BIN" ]]; then
  for candidate in /opt/google-cloud-sdk/bin/gsutil /root/google-cloud-sdk/bin/gsutil; do
    if [[ -x "$candidate" ]]; then
      GSUTIL_BIN="$candidate"
      break
    fi
  done
fi
if [[ -z "$GSUTIL_BIN" ]]; then
  echo "❌ gsutil not found. Checked PATH and known Cloud SDK install locations."
  echo "   PATH=$PATH"
  exit 1
fi
echo "✅ Using gsutil binary: $GSUTIL_BIN"

sanitize_filename() {
  echo "$1" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9._-]/_/g; s/__*/_/g; s/^_//; s/_$//'
}

# 0. System diagnostics
echo "🔍 System Memory Information:"
free -h
echo "💻 Available CPU cores: $(nproc)"
echo "☕ Java version:"
java -version 2>&1
echo "🏃 JVM memory settings: $JAVA_TOOL_OPTIONS"
echo ""

# 1. Validate crawl type
if [[ "$CRAWL_TYPE" != "daily" && "$CRAWL_TYPE" != "weekly" && "$CRAWL_TYPE" != "custom" ]]; then
  echo "❌ CRAWL_TYPE must be 'daily', 'weekly', or 'custom'"
  exit 1
fi

# 1.1. Validate custom crawl requirements
if [[ "$CRAWL_TYPE" == "custom" ]]; then
  if [[ -z "$CUSTOM_URL" ]]; then
    echo "❌ CUSTOM_URL is required for custom crawl type"
    exit 1
  fi
  if [[ -z "$JOB_NAME" ]]; then
    echo "❌ JOB_NAME is required for custom crawl type"
    exit 1
  fi
  echo "🎯 Custom crawl mode: $CUSTOM_URL"
elif [[ -z "$JOB_NAME" ]]; then
  # For scheduled crawls, default the job name to crawl type so GCS files resolve.
  JOB_NAME="$CRAWL_TYPE"
  echo "ℹ️ JOB_NAME not provided. Defaulting to: $JOB_NAME"
fi

# 2. Set paths
DATE=$(date +%Y-%m-%d)
GCS_BUCKET="gs://screamingfrog-configs"
CONFIG_FILE="${JOB_NAME}.seospiderconfig"
DOMAINS_FILE="${JOB_NAME}_domains.txt"
OUTPUT_PATH="/app/output/$DATE/$JOB_NAME"
COMBINED_DIR="$OUTPUT_PATH/combined_reports"
INDEX_FILE="$COMBINED_DIR/report_index.csv"
README_FILE="$COMBINED_DIR/README.txt"

mkdir -p "$OUTPUT_PATH"
mkdir -p /app/configs

# 3. Fetch license
export SF_DEBUG_LICENSE=true
mkdir -p ~/.ScreamingFrogSEOSpider

echo "🔐 Setting up license..."
echo "License username length: ${#SF_LICENSE_USERNAME}"
echo "License key length: ${#SF_LICENSE_KEY}"

cat <<EOF > ~/.ScreamingFrogSEOSpider/licence.txt
${SF_LICENSE_USERNAME}
${SF_LICENSE_KEY}
EOF

echo "✅ License file created at: ~/.ScreamingFrogSEOSpider/licence.txt"
echo "License file contents (first 50 chars):"
head -c 50 ~/.ScreamingFrogSEOSpider/licence.txt
echo ""

# 4. Initialize gcloud and download config + domain list from GCS
echo "🔧 Checking authentication..."
echo "Current service account:"
gcloud auth list --filter=status:ACTIVE --format="value(account)" 2>/dev/null || echo "No active account found"

echo "🔐 Authenticating with service account..."

if [[ "$CRAWL_TYPE" == "custom" ]]; then
  echo "📥 Using bundled custom crawl config..."
  if [[ ! -f "/app/configs/activated.seospiderconfig" ]]; then
    echo "❌ Missing config: /app/configs/activated.seospiderconfig"
    exit 1
  fi

  # Create a temporary domains file with just the custom URL.
  echo "$CUSTOM_URL" > "/app/configs/domains.txt"
  echo "✅ Config: /app/configs/activated.seospiderconfig"
  echo "✅ Custom URL: $CUSTOM_URL"
else
  echo "📥 Downloading crawl config and domain list from GCS..."
  "$GSUTIL_BIN" cp "$GCS_BUCKET/$CONFIG_FILE" "/app/configs/activated.seospiderconfig"
  "$GSUTIL_BIN" cp "$GCS_BUCKET/$DOMAINS_FILE" "/app/configs/domains.txt"
  echo "✅ Config: $CONFIG_FILE"
  echo "✅ Domains: $DOMAINS_FILE"
fi

EXPORT_TABS=(
  "Internal:All"
  "Page Titles:Missing"
  "Page Titles:Duplicate"
  "Page Titles:Same as H1"
  "Page Titles:Multiple"
  "Page Titles:Outside <head>"
  "Meta Description:Missing"
  "Meta Description:Duplicate"
  "Meta Description:Below X Characters"
  "Meta Description:Multiple"
  "Meta Description:Outside <head>"
  "H1:Missing"
  "H1:Duplicate"
  "H1:Multiple"
  "H2:Missing"
  "H2:Non-Sequential"
  "JavaScript:Contains JavaScript Content"
  "JavaScript:Noindex Only in Original HTML"
  "JavaScript:Nofollow Only in Original HTML"
  "JavaScript:Canonical Only in Rendered HTML"
  "JavaScript:Canonical Mismatch"
  "Response Codes:Internal Redirect Chain"
  "Response Codes:Internal Redirect Loop"
  "Structured Data:Validation Errors"
  "Structured Data:Validation Warnings"
  "Sitemaps:Orphan URLs"
  "Sitemaps:Non-indexable URLs in Sitemap"
  "Sitemaps:URLs not in Sitemap"
  "URL:Underscores"
  "URL:Uppercase"
  "URL:Multiple Slashes"
  "URL:Repetitive Path"
  "URL:Contains Space"
  "URL:Over X Characters"
  "URL:Parameters"
)
EXPORT_TABS_ARG=$(IFS=,; echo "${EXPORT_TABS[*]}")

BULK_EXPORTS=(
  "All Inlinks"
  "Response Codes:Client Error (4xx) Inlinks"
  "Response Codes:Redirection (3xx) Inlinks"
  "Response Codes:Server Error (5xx) Inlinks"
  "Response Codes:Blocked by Robots.txt Inlinks"
  "Response Codes:Blocked Resource Inlinks"
  "Response Codes:No Response Inlinks"
  "Images:Images Missing Alt Text Inlinks"
  "Canonicals:Missing Inlinks"
  "Canonicals:Non-Indexable Canonical Inlinks"
  "Canonicals:Multiple Inlinks"
  "Directives:Noindex Inlinks"
  "Directives:Nofollow Inlinks"
  "Content:Soft 404 Inlinks"
)
BULK_EXPORTS_ARG=$(IFS=,; echo "${BULK_EXPORTS[*]}")

# 5. Crawl each domain
echo "🚀 Starting crawl..."
while IFS= read -r DOMAIN || [[ -n "$DOMAIN" ]]; do
  # Skip empty lines and comments.
  if [[ -z "$DOMAIN" || "$DOMAIN" =~ ^[[:space:]]*$ || "$DOMAIN" =~ ^[[:space:]]*# ]]; then
    continue
  fi

  # Trim whitespace.
  DOMAIN=$(echo "$DOMAIN" | xargs)

  # Create domain-specific output folder.
  DOMAIN_CLEAN=$(echo "$DOMAIN" | sed 's/[^a-zA-Z0-9.-]/_/g')
  DOMAIN_OUTPUT="$OUTPUT_PATH/$DOMAIN_CLEAN"
  mkdir -p "$DOMAIN_OUTPUT"

  echo "🌐 Crawling $DOMAIN → $DOMAIN_OUTPUT"
  echo "🔍 Pre-crawl memory status:"
  free -h

  /usr/bin/screamingfrogseospider \
    --headless \
    --config /app/configs/activated.seospiderconfig \
    --crawl "$DOMAIN" \
    --output-folder "$DOMAIN_OUTPUT" \
    --overwrite \
    --export-format csv \
    --export-tabs "$EXPORT_TABS_ARG" \
    --bulk-export "$BULK_EXPORTS_ARG"
done < /app/configs/domains.txt

# 6. Combine CSVs by report name so every output keeps the correct columns.
echo "📦 Combining CSVs by report type..."
rm -rf "$COMBINED_DIR"
mkdir -p "$COMBINED_DIR"

printf 'Report Name,Combined File,Domain Count,Row Count\n' > "$INDEX_FILE"

cat <<'EOF' > "$README_FILE"
These exports are combined by report type.

- Each CSV in this folder contains only one Screaming Frog report, combined across all crawled domains.
- Two helper columns were added at the front of each file: Domain and Source File.
- Use internal_all.csv when you want the broadest page-level view.
- Use h1_all.csv, h2_all.csv, page_titles_all.csv, and meta_description_all.csv for focused on-page analysis.
- Use report_index.csv as the table of contents for this archive.
EOF

declare -A REPORT_NAMES=()

shopt -s nullglob
for domain_folder in "$OUTPUT_PATH"/*/; do
  [[ -d "$domain_folder" ]] || continue

  DOMAIN_NAME=$(basename "$domain_folder")
  [[ "$DOMAIN_NAME" == "$(basename "$COMBINED_DIR")" ]] && continue
  echo "🏷️  Processing domain: $DOMAIN_NAME"

  for file in "$domain_folder"/*.csv; do
    [[ -f "$file" ]] || continue

    report_file=$(basename "$file")
    report_name="${report_file%.csv}"
    safe_report_name=$(sanitize_filename "$report_name")
    combined_file="$COMBINED_DIR/${safe_report_name}.csv"

    echo "📄 Merging report: $report_file"

    if [[ ! -f "$combined_file" ]]; then
      printf 'Domain,Source File,' > "$combined_file"
      head -n 1 "$file" >> "$combined_file"
      REPORT_NAMES["$safe_report_name"]="$report_name"
    fi

    while IFS= read -r line || [[ -n "$line" ]]; do
      if [[ -n "$line" ]]; then
        printf '%s,%s,%s\n' "$DOMAIN_NAME" "$report_file" "$line" >> "$combined_file"
      fi
    done < <(tail -n +2 "$file")
  done
done
shopt -u nullglob

echo "📚 Building report index..."
shopt -s nullglob
for combined_file in "$COMBINED_DIR"/*.csv; do
  [[ "$(basename "$combined_file")" == "report_index.csv" ]] && continue

  combined_name=$(basename "$combined_file")
  safe_report_name="${combined_name%.csv}"
  report_name="${REPORT_NAMES[$safe_report_name]:-$safe_report_name}"
  row_count=$(( $(wc -l < "$combined_file") - 1 ))
  domain_count=$(tail -n +2 "$combined_file" | cut -d',' -f1 | sort -u | wc -l)

  printf '"%s","%s",%s,%s\n' \
    "$report_name" \
    "$combined_name" \
    "$domain_count" \
    "$row_count" >> "$INDEX_FILE"
done
shopt -u nullglob

echo "✅ Clear report exports created in: $COMBINED_DIR"
echo "📘 Report index: $INDEX_FILE"
echo "📄 Readme: $README_FILE"

# 6.1 Package the combined reports into a single archive for one-click download.
ARCHIVE_LABEL="${JOB_NAME}_${DATE}_reports"
ARCHIVE_PATH=""

if command -v zip >/dev/null 2>&1; then
  ARCHIVE_PATH="$OUTPUT_PATH/${ARCHIVE_LABEL}.zip"
  echo "🗜️ Creating zip archive: $ARCHIVE_PATH"
  (
    cd "$OUTPUT_PATH"
    zip -qr "$ARCHIVE_PATH" "$(basename "$COMBINED_DIR")"
  )
else
  ARCHIVE_PATH="$OUTPUT_PATH/${ARCHIVE_LABEL}.tar.gz"
  echo "🗜️ zip not found, creating tar.gz archive instead: $ARCHIVE_PATH"
  tar -C "$OUTPUT_PATH" -czf "$ARCHIVE_PATH" "$(basename "$COMBINED_DIR")"
fi

if [[ ! -f "$ARCHIVE_PATH" ]]; then
  echo "❌ Failed to create reports archive."
  exit 1
fi

# 7. Upload to Google Cloud Storage
ARCHIVE_BASENAME=$(basename "$ARCHIVE_PATH")
GCS_UPLOAD_PATH="screamingfrog-reports/$CRAWL_TYPE/$DATE/$JOB_NAME/$ARCHIVE_BASENAME"
SIGNED_URL_EXPIRY=$((60 * 60 * 24 * 7))

echo "📤 Uploading reports archive to GCS..."
if "$GSUTIL_BIN" cp "$ARCHIVE_PATH" "$GCS_BUCKET/$GCS_UPLOAD_PATH"; then
  echo "✅ Upload successful to $GCS_BUCKET/$GCS_UPLOAD_PATH"
else
  echo "❌ GCS upload failed!"
  exit 1
fi

echo "🔗 Generating signed URL (valid for 7 days)..."

if [ ! -f "/app/service-account.json" ]; then
  echo "❌ Service account file not found at /app/service-account.json"
  echo "📂 Checking available files in /app:"
  ls -la /app/*.json 2>/dev/null || echo "No JSON files found in /app"
  SIGNED_URL="FILE_NOT_FOUND"
else
  echo "🔍 Using service account: /app/service-account.json"
  echo "🎯 Generating signed URL for: $GCS_BUCKET/$GCS_UPLOAD_PATH"

  SIGNURL_OUTPUT=$("$GSUTIL_BIN" signurl -d ${SIGNED_URL_EXPIRY}s /app/service-account.json "$GCS_BUCKET/$GCS_UPLOAD_PATH" 2>&1)
  SIGNURL_EXIT_CODE=$?

  if [ $SIGNURL_EXIT_CODE -ne 0 ]; then
    echo "❌ Failed to generate signed URL!"
    echo "🔍 Exit code: $SIGNURL_EXIT_CODE"
    echo "🔍 Error output:"
    echo "$SIGNURL_OUTPUT"
    echo "🔍 Command attempted: $GSUTIL_BIN signurl -d ${SIGNED_URL_EXPIRY}s /app/service-account.json $GCS_BUCKET/$GCS_UPLOAD_PATH"
    echo "📂 Service account file info:"
    ls -la /app/service-account.json 2>/dev/null || echo "Service account file not found!"
    echo "🔐 Current gcloud auth status:"
    gcloud auth list --filter=status:ACTIVE --format="value(account)" 2>/dev/null || echo "No active gcloud auth found"
    SIGNED_URL="GENERATION_FAILED"
  else
    SIGNED_URL=$(echo "$SIGNURL_OUTPUT" | tail -n 1 | awk '{print $NF}')
    echo "✅ Signed URL generated successfully"
  fi
fi

echo "✅ Signed URL result:"
echo "$SIGNED_URL"

# Save for Slack or other integrations.
echo "$SIGNED_URL" > "$OUTPUT_PATH/signed_url.txt"

# 7.1 Send signed URL to Slack
echo "📨 Sending signed URL to Slack..."

if [[ "$CRAWL_TYPE" == "custom" ]]; then
  SLACK_MESSAGE="🚀 *Custom Screaming Frog SEO Crawl Complete*\n📅 *Date:* $DATE\n🌐 *URL:* $CUSTOM_URL\n🔗 *Job Name:* $JOB_NAME\n📦 *Archive:* <$SIGNED_URL|Download Reports Archive>\n(valid for 7 days)"
else
  SLACK_MESSAGE="🚀 *Screaming Frog SEO Crawl Complete*\n📅 *Date:* $DATE\n🗂️ *Type:* $CRAWL_TYPE\n🔗 *Job Name:* $JOB_NAME\n📦 *Archive:* <$SIGNED_URL|Download Reports Archive>\n(valid for 7 days)"
fi

SLACK_RESPONSE=$(curl -sS -X POST -H 'Content-type: application/json' \
  --data "{\"text\": \"$SLACK_MESSAGE\"}" \
  "$SLACK_WEBHOOK_URL" || true)

if [[ "$SLACK_RESPONSE" == "ok" ]]; then
  echo "✅ Slack notification sent successfully (ok%)."
elif [[ "$SLACK_RESPONSE" == "no_service" ]]; then
  echo "❌ Slack notification failed (no_service%). Check webhook URL/service."
  exit 1
else
  echo "❌ Slack notification failed. Response: $SLACK_RESPONSE"
  exit 1
fi

# 8. Clean up
echo "🧹 Cleaning up output folder..."
find /app/output/ -mindepth 1 -delete

echo "✅ Done!"
