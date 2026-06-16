use chrono::{Offset as _, TimeZone as _};
use time::{Date, OffsetDateTime, PrimitiveDateTime, Time, UtcOffset};

const APP_TIMEZONE_VAR: &str = "DASHBOARD_TIMEZONE";
const DB_TIMEZONE_VAR: &str = "DASHBOARD_DB_TIMEZONE";
const SYSTEM_TIMEZONE_VAR: &str = "TZ";
const DEFAULT_TIMEZONE: &str = "Asia/Shanghai";

/// Returns the current local time using chrono-tz instead of the platform localtime APIs.
pub fn now_local() -> OffsetDateTime {
    let utc_now = chrono::Utc::now();
    let offset = offset_for_unix_timestamp(utc_now.timestamp());
    let nanos = i128::from(utc_now.timestamp()) * 1_000_000_000
        + i128::from(utc_now.timestamp_subsec_nanos());

    OffsetDateTime::from_unix_timestamp_nanos(nanos)
        .unwrap_or_else(|_| OffsetDateTime::now_utc())
        .to_offset(offset)
}

/// Returns the current Unix timestamp in milliseconds using the configured local timezone.
pub fn now_millis() -> i64 {
    now_local().unix_timestamp_nanos() as i64 / 1_000_000
}

/// Returns the current configured local UTC offset.
pub fn local_offset() -> UtcOffset {
    offset_for_unix_timestamp(chrono::Utc::now().timestamp())
}

/// Returns the configured local UTC offset in milliseconds for SQL wall-clock reconstruction.
pub fn local_offset_millis() -> i64 {
    i64::from(local_offset().whole_seconds()) * 1000
}

/// Returns the Unix timestamp in milliseconds for the start of a configured local calendar day.
pub fn local_day_start_millis(date: Date) -> i64 {
    PrimitiveDateTime::new(date, Time::MIDNIGHT)
        .assume_offset(offset_for_local_midnight(date))
        .unix_timestamp()
        * 1000
}

/// Chooses the configured IANA timezone name without requiring global unsafe time crate setup.
fn configured_timezone_name() -> String {
    [APP_TIMEZONE_VAR, DB_TIMEZONE_VAR, SYSTEM_TIMEZONE_VAR]
        .into_iter()
        .find_map(|key| {
            std::env::var(key)
                .ok()
                .filter(|value| !value.trim().is_empty())
        })
        .unwrap_or_else(|| DEFAULT_TIMEZONE.to_string())
}

/// Parses the configured timezone and falls back to UTC for invalid deployment configuration.
fn configured_timezone() -> chrono_tz::Tz {
    configured_timezone_name()
        .parse::<chrono_tz::Tz>()
        .unwrap_or(chrono_tz::UTC)
}

/// Resolves the offset for a specific instant, so DST transitions are handled by chrono-tz data.
fn offset_for_unix_timestamp(timestamp: i64) -> UtcOffset {
    let local = configured_timezone().timestamp_opt(timestamp, 0);
    let offset_seconds = local
        .single()
        .map(|dt| dt.offset().fix().local_minus_utc())
        .unwrap_or(0);

    offset_from_seconds(offset_seconds)
}

/// Resolves midnight with the earliest valid local instant when timezone rules create ambiguity.
fn offset_for_local_midnight(date: Date) -> UtcOffset {
    let local = configured_timezone().with_ymd_and_hms(
        date.year(),
        date.month() as u32,
        date.day() as u32,
        0,
        0,
        0,
    );
    let offset_seconds = local
        .earliest()
        .map(|dt| dt.offset().fix().local_minus_utc())
        .unwrap_or(0);

    offset_from_seconds(offset_seconds)
}

/// Converts chrono's offset seconds into time's offset type while keeping invalid data contained.
fn offset_from_seconds(offset_seconds: i32) -> UtcOffset {
    UtcOffset::from_whole_seconds(offset_seconds).unwrap_or(UtcOffset::UTC)
}
