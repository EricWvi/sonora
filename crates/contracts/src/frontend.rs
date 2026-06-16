use serde::Serialize;

/// Enumerates the HTTP methods supported by the generated frontend SDK.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "UPPERCASE")]
pub enum FrontendHttpMethod {
    Get,
    Post,
    Put,
    Delete,
}

/// Describes one request field that the transport must interpolate into the URL path.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct FrontendPathParam {
    pub rust_field_name: &'static str,
    pub wire_name: &'static str,
}

/// Describes one frontend-facing HTTP operation exported from `sonora-contracts`.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct FrontendEndpoint {
    pub operation_name: &'static str,
    pub method: FrontendHttpMethod,
    pub path_template: &'static str,
    pub request_type: &'static str,
    pub response_type: &'static str,
    pub path_params: &'static [FrontendPathParam],
    pub has_json_body: bool,
}

const FRONTEND_ENDPOINTS: &[FrontendEndpoint] = &[];

/// Returns the Rust-owned endpoint metadata exported to the generated frontend SDK.
pub fn frontend_endpoints() -> &'static [FrontendEndpoint] {
    FRONTEND_ENDPOINTS
}
