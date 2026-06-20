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

/// Describes one request field that the transport must either interpolate into the URL path
/// or append as a query string parameter.
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
    /// TypeScript type for the request object passed by the caller.
    ///
    /// Named contract types (e.g. `CreateDirRequest`) are imported from their DTO module.
    /// Inline object types (e.g. `{ id: string }`) and primitives (`boolean`, `void`) need
    /// no import and are written verbatim into the generated SDK.
    pub request_type: &'static str,
    /// TypeScript type for the resolved response value.
    pub response_type: &'static str,
    /// Fields extracted from the request object and interpolated into `path_template`.
    pub path_params: &'static [FrontendPathParam],
    /// Fields extracted from the request object and appended as URL query parameters.
    pub query_params: &'static [FrontendPathParam],
    pub has_json_body: bool,
}

const NODE_ID_PARAM: FrontendPathParam = FrontendPathParam {
    rust_field_name: "id",
    wire_name: "id",
};

const PATH_QUERY_PARAM: FrontendPathParam = FrontendPathParam {
    rust_field_name: "path",
    wire_name: "path",
};

const FRONTEND_ENDPOINTS: &[FrontendEndpoint] = &[
    FrontendEndpoint {
        operation_name: "createDir",
        method: FrontendHttpMethod::Post,
        path_template: "/api/nodes/dirs",
        request_type: "CreateDirRequest",
        response_type: "CreateDirResponse",
        path_params: &[],
        query_params: &[],
        has_json_body: true,
    },
    FrontendEndpoint {
        operation_name: "createFile",
        method: FrontendHttpMethod::Post,
        path_template: "/api/nodes/files",
        request_type: "CreateFileRequest",
        response_type: "CreateFileResponse",
        path_params: &[],
        query_params: &[],
        has_json_body: true,
    },
    FrontendEndpoint {
        operation_name: "getNodeByPath",
        method: FrontendHttpMethod::Get,
        path_template: "/api/nodes/by-path",
        request_type: "GetNodeByPathRequest",
        response_type: "GetNodeResponse",
        path_params: &[],
        query_params: &[PATH_QUERY_PARAM],
        has_json_body: false,
    },
    FrontendEndpoint {
        operation_name: "nodeExists",
        method: FrontendHttpMethod::Get,
        path_template: "/api/nodes/exists",
        request_type: "NodeExistsRequest",
        response_type: "boolean",
        path_params: &[],
        query_params: &[PATH_QUERY_PARAM],
        has_json_body: false,
    },
    FrontendEndpoint {
        operation_name: "listRootChildren",
        method: FrontendHttpMethod::Get,
        path_template: "/api/nodes/root/children",
        request_type: "ListRootChildrenRequest",
        response_type: "ListChildrenResponse",
        path_params: &[],
        query_params: &[],
        has_json_body: false,
    },
    FrontendEndpoint {
        operation_name: "getNodeById",
        method: FrontendHttpMethod::Get,
        path_template: "/api/nodes/{id}",
        request_type: "GetNodeByIdRequest",
        response_type: "GetNodeResponse",
        path_params: &[NODE_ID_PARAM],
        query_params: &[],
        has_json_body: false,
    },
    FrontendEndpoint {
        operation_name: "listChildren",
        method: FrontendHttpMethod::Get,
        path_template: "/api/nodes/{id}/children",
        request_type: "ListChildrenRequest",
        response_type: "ListChildrenResponse",
        path_params: &[NODE_ID_PARAM],
        query_params: &[],
        has_json_body: false,
    },
    FrontendEndpoint {
        operation_name: "getNodePath",
        method: FrontendHttpMethod::Get,
        path_template: "/api/nodes/{id}/path",
        request_type: "GetNodePathRequest",
        response_type: "GetPathResponse",
        path_params: &[NODE_ID_PARAM],
        query_params: &[],
        has_json_body: false,
    },
    FrontendEndpoint {
        operation_name: "moveNode",
        method: FrontendHttpMethod::Put,
        path_template: "/api/nodes/{id}/move",
        request_type: "MoveNodeWithIdRequest",
        response_type: "MoveNodeResponse",
        path_params: &[NODE_ID_PARAM],
        query_params: &[],
        has_json_body: true,
    },
    FrontendEndpoint {
        operation_name: "deleteNode",
        method: FrontendHttpMethod::Delete,
        path_template: "/api/nodes/{id}",
        request_type: "DeleteNodeRequest",
        response_type: "null",
        path_params: &[NODE_ID_PARAM],
        query_params: &[],
        has_json_body: false,
    },
];

/// Returns the Rust-owned endpoint metadata exported to the generated frontend SDK.
pub fn frontend_endpoints() -> &'static [FrontendEndpoint] {
    FRONTEND_ENDPOINTS
}
