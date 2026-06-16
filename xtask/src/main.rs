use std::process::ExitCode;

/// Runs the requested xtask command from the workspace root.
fn main() -> ExitCode {
    match run() {
        Ok(()) => ExitCode::SUCCESS,
        Err(error) => {
            eprintln!("{error}");
            ExitCode::FAILURE
        }
    }
}

/// Parses the xtask command line and dispatches to the matching workflow.
fn run() -> Result<(), String> {
    let mut arguments = std::env::args().skip(1);
    let Some(command) = arguments.next() else {
        return Err("usage: cargo xtask export-contracts".to_string());
    };

    if command != "export-contracts" {
        return Err(format!("unknown xtask command `{command}`"));
    }

    if let Some(unexpected) = arguments.next() {
        return Err(format!("unexpected argument `{unexpected}`"));
    }

    let workspace_root = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
        .parent()
        .ok_or_else(|| "failed to determine workspace root".to_string())?;

    sonora_xtask::run_export_contracts(workspace_root)
        .map_err(|error| format!("failed to export contracts: {error}"))
}
