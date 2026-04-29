function lines(value) {
  return Array.isArray(value) ? value.join('\n') : ''
}

function list(value) {
  return String(value ?? '')
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
}

export function auditDefaultsToForm(config = {}) {
  return {
    url: config.url ?? '',
    scope: config.scope ?? 'host',
    scope_prefix: config.scope_prefix ?? '',
    sitemap_url: config.sitemap_url ?? '',
    sitemap_mode: config.sitemap_mode ?? 'off',
    max_depth: config.max_depth ?? -1,
    max_pages: config.max_pages ?? 0,
    concurrency: config.concurrency ?? 10,
    timeout: config.timeout ?? '30s',
    platform: config.platform ?? '',
    user_agent: config.user_agent ?? '',
    mobile_user_agent: config.mobile_user_agent ?? '',
    respect_robots: config.respect_robots ?? true,
    max_redirects: config.max_redirects ?? 10,
    max_page_size_kb: config.max_page_size_kb ?? 5120,
    max_url_length: config.max_url_length ?? 0,
    max_query_params: config.max_query_params ?? 0,
    max_links_per_page: config.max_links_per_page ?? 0,
    follow_nofollow_links: config.follow_nofollow_links ?? false,
    expand_noindex_pages: config.expand_noindex_pages ?? true,
    expand_canonicalized_pages: config.expand_canonicalized_pages ?? true,
    output_dir: config.output_dir ?? '',
    validate_external_links: config.validate_external_links ?? true,
    discover_resources: config.discover_resources ?? true,
    enable_crawler_evidence: config.enable_crawler_evidence ?? true,
    enable_rendered_seo: config.enable_rendered_seo ?? true,
    rendered_sample_limit: config.rendered_sample_limit ?? 5,
    rendered_timeout: config.rendered_timeout ?? '20s',
    expected_inventory_urls: lines(config.expected_inventory_urls),
    expected_parameter_names: lines(config.expected_parameter_names),
    allowed_image_cdn_hosts: lines(config.allowed_image_cdn_hosts),
    required_live_text: lines(config.required_live_text),
  }
}

export function auditFormToRequest(form) {
  return {
    ...form,
    max_redirects: Number(form.max_redirects),
    max_page_size_kb: Number(form.max_page_size_kb),
    max_url_length: Number(form.max_url_length),
    max_query_params: Number(form.max_query_params),
    max_links_per_page: Number(form.max_links_per_page),
    max_depth: Number(form.max_depth),
    max_pages: Number(form.max_pages),
    concurrency: Number(form.concurrency),
    rendered_sample_limit: Number(form.rendered_sample_limit),
    expected_inventory_urls: list(form.expected_inventory_urls),
    expected_parameter_names: list(form.expected_parameter_names),
    allowed_image_cdn_hosts: list(form.allowed_image_cdn_hosts),
    required_live_text: list(form.required_live_text),
  }
}

export function optionsOrCurrent(options, value, fallbackLabel) {
  return options?.length ? options : [{ value, label: fallbackLabel ?? String(value) }]
}
