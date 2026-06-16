use tracing_subscriber::Registry;
use tracing_subscriber::filter::LevelFilter;
use tracing_subscriber::layer::Layer;
use tracing_subscriber::prelude::*;

/// Runs one test body under a TRACE-enabled subscriber so shared callsites stay observable.
pub fn with_trace_logging<R>(action: impl FnOnce() -> R) -> R {
    let subscriber = tracing_subscriber::registry().with(LevelFilter::TRACE);

    tracing::subscriber::with_default(subscriber, action)
}

/// Installs a TRACE-enabled subscriber as the thread default and returns a guard.
///
/// The subscriber remains active until the guard is dropped. Use this form when
/// the test body is `async` and therefore cannot be wrapped in a sync closure.
/// For synchronous test bodies, prefer [`with_trace_logging`].
pub fn set_trace_logging() -> tracing::subscriber::DefaultGuard {
    let subscriber = tracing_subscriber::registry().with(LevelFilter::TRACE);
    tracing::subscriber::set_default(subscriber)
}

/// Runs one test body under a TRACE-enabled subscriber that records structured logging events.
pub fn with_recorded_trace_logging<L, R>(layer: L, action: impl FnOnce() -> R) -> R
where
    L: Layer<Registry> + Send + Sync + 'static,
{
    let subscriber = tracing_subscriber::registry()
        .with(layer)
        .with(LevelFilter::TRACE);

    tracing::subscriber::with_default(subscriber, action)
}
