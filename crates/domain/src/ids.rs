use std::fmt;

use uuid::Uuid;

/// Generates a type-safe newtype ID backed by a [`Uuid`].
///
/// The generated struct derives `Clone`, `Copy`, `Debug`, `Eq`, `Hash`, `Ord`, `PartialEq`,
/// `PartialOrd`, and `Display`. With the `serde` feature it serializes as the UUID value.
/// With the `sqlx` feature it implements `Encode`, `Decode`, and `Type` for PostgreSQL,
/// delegating directly to the `Uuid` implementations.
macro_rules! define_uuid_id {
    ($name:ident, $doc:literal) => {
        #[doc = $doc]
        #[derive(Clone, Copy, Debug, Eq, Hash, Ord, PartialEq, PartialOrd)]
        #[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
        pub struct $name(Uuid);

        impl $name {
            /// Wraps an existing UUID value.
            pub fn new(uuid: Uuid) -> Self {
                Self(uuid)
            }

            /// Generates a new random ID.
            pub fn generate() -> Self {
                Self(Uuid::new_v4())
            }

            /// Parses an ID from its UUID string representation.
            pub fn parse(s: &str) -> Result<Self, uuid::Error> {
                s.parse::<Uuid>().map(Self::new)
            }

            /// Returns the underlying UUID value.
            pub fn as_uuid(self) -> Uuid {
                self.0
            }
        }

        impl fmt::Display for $name {
            fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
                write!(f, "{}", self.0)
            }
        }

        #[cfg(feature = "sqlx")]
        impl<'q> sqlx::Encode<'q, sqlx::Postgres> for $name {
            fn encode_by_ref(
                &self,
                buf: &mut <sqlx::Postgres as sqlx::Database>::ArgumentBuffer,
            ) -> Result<sqlx::encode::IsNull, sqlx::error::BoxDynError> {
                <Uuid as sqlx::Encode<sqlx::Postgres>>::encode_by_ref(&self.0, buf)
            }
        }

        #[cfg(feature = "sqlx")]
        impl<'r> sqlx::Decode<'r, sqlx::Postgres> for $name {
            fn decode(
                value: <sqlx::Postgres as sqlx::Database>::ValueRef<'r>,
            ) -> Result<Self, sqlx::error::BoxDynError> {
                let uuid = <Uuid as sqlx::Decode<sqlx::Postgres>>::decode(value)?;
                Ok(Self(uuid))
            }
        }

        #[cfg(feature = "sqlx")]
        impl sqlx::Type<sqlx::Postgres> for $name {
            fn type_info() -> <sqlx::Postgres as sqlx::Database>::TypeInfo {
                <Uuid as sqlx::Type<sqlx::Postgres>>::type_info()
            }
        }
    };
}

/// Generates a type-safe newtype ID backed by an integer primitive.
///
/// The generated struct derives `Clone`, `Copy`, `Debug`, `Eq`, `Hash`, `Ord`, `PartialEq`,
/// `PartialOrd`, and `Display`. With the `serde` feature it serializes as the integer value.
/// With the `sqlx` feature it implements `Encode`, `Decode`, and `Type` for PostgreSQL,
/// delegating to the inner primitive type's implementations.
#[allow(unused_macros)]
macro_rules! define_int_id {
    ($name:ident, $inner:ty, $doc:literal) => {
        #[doc = $doc]
        #[derive(Clone, Copy, Debug, Eq, Hash, Ord, PartialEq, PartialOrd)]
        #[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
        pub struct $name($inner);

        impl $name {
            /// Wraps an existing integer value.
            pub fn new(value: $inner) -> Self {
                Self(value)
            }

            /// Returns the underlying integer value.
            pub fn value(self) -> $inner {
                self.0
            }
        }

        impl fmt::Display for $name {
            fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
                write!(f, "{}", self.0)
            }
        }

        #[cfg(feature = "sqlx")]
        impl<'q> sqlx::Encode<'q, sqlx::Postgres> for $name {
            fn encode_by_ref(
                &self,
                buf: &mut <sqlx::Postgres as sqlx::Database>::ArgumentBuffer,
            ) -> Result<sqlx::encode::IsNull, sqlx::error::BoxDynError> {
                <$inner as sqlx::Encode<sqlx::Postgres>>::encode_by_ref(&self.0, buf)
            }
        }

        #[cfg(feature = "sqlx")]
        impl<'r> sqlx::Decode<'r, sqlx::Postgres> for $name {
            fn decode(
                value: <sqlx::Postgres as sqlx::Database>::ValueRef<'r>,
            ) -> Result<Self, sqlx::error::BoxDynError> {
                let v = <$inner as sqlx::Decode<sqlx::Postgres>>::decode(value)?;
                Ok(Self(v))
            }
        }

        #[cfg(feature = "sqlx")]
        impl sqlx::Type<sqlx::Postgres> for $name {
            fn type_info() -> <sqlx::Postgres as sqlx::Database>::TypeInfo {
                <$inner as sqlx::Type<sqlx::Postgres>>::type_info()
            }
        }
    };
}

define_uuid_id!(NodeId, "Identifies a VFS node.");
