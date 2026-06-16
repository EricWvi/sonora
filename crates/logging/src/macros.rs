/// Emits a debug event and automatically records the current function name under `method`.
#[macro_export]
macro_rules! sonora_debug {
    ($($arg:tt)*) => {
        {
            fn __sonora_log_marker() {}
            ::tracing::debug!(
                method = $crate::__private::method_name_from_marker_type_name(
                    ::core::any::type_name_of_val(&__sonora_log_marker),
                    "__sonora_log_marker",
                ),
                $($arg)*
            )
        }
    };
}

/// Emits an info event and automatically records the current function name under `method`.
#[macro_export]
macro_rules! sonora_info {
    ($($arg:tt)*) => {
        {
            fn __sonora_log_marker() {}
            ::tracing::info!(
                method = $crate::__private::method_name_from_marker_type_name(
                    ::core::any::type_name_of_val(&__sonora_log_marker),
                    "__sonora_log_marker",
                ),
                $($arg)*
            )
        }
    };
}

/// Emits a warning event and automatically records the current function name under `method`.
#[macro_export]
macro_rules! sonora_warn {
    ($($arg:tt)*) => {
        {
            fn __sonora_log_marker() {}
            ::tracing::warn!(
                method = $crate::__private::method_name_from_marker_type_name(
                    ::core::any::type_name_of_val(&__sonora_log_marker),
                    "__sonora_log_marker",
                ),
                $($arg)*
            )
        }
    };
}

/// Emits an error event and automatically records the current function name under `method`.
#[macro_export]
macro_rules! sonora_error {
    ($($arg:tt)*) => {
        {
            fn __sonora_log_marker() {}
            ::tracing::error!(
                method = $crate::__private::method_name_from_marker_type_name(
                    ::core::any::type_name_of_val(&__sonora_log_marker),
                    "__sonora_log_marker",
                ),
                $($arg)*
            )
        }
    };
}
