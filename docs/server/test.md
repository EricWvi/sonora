## Testing Considerations

When writing integrated tests for handlers, be aware of the following requirements:

### 1. Init Stage
- Tests must call `config.Init()` before using database or storage services
- Tests should use `config.TestRouter()` to get router.

### 2. Action Middleware for Tests
- **Issue**: The handler dispatch system requires the `Action` parameter to be set in the gin context via `c.Set("Action", action)`
- **Production**: The `middleware.Logging()` middleware extracts Action from query parameters and sets it
- **Test Solution**: Add a simple middleware to test routers:
```go
router.Use(func(c *gin.Context) {
    action := c.Request.URL.Query().Get("Action")
    if action != "" {
        c.Set("Action", action)
    }
    c.Next()
})
```

### 3. Test Structure for Operations
- always create new records to test update/delete
- use `defer` to clean up created records
- if an api helper does not exist, say `UpdateTestAlbum`, add it in `tests/api/`.

### 4. Database Access in Tests
- Use `config.DB()` to get the database instance in tests
- The database is shared across test runs - ensure proper cleanup or use unique test data
- Tests use soft deletes (GORM), so deleted records have `deleted_at` timestamp set
