`sonora` is a self-hosting music streaming service. It comprises two parts, admin management system and music player.

The backend is designed around models and actions on the models. Database tables are in `migration/migrations.go`. Models are defined in `model`.

The overall HTTP interface is through a `base` handler and using `Action` query and reflection to choose method on `base`. For each group of handlers, say `singer`, we will register it in `router.go`, using `GET` and `POST`. The corresponding handlers for actions of a specific model, say `singer`, are defined in `handler/<model>`, say `handler/singer`.

The current development is in admin part.

## Database Schema (from migration/migrations.go)
- **d_singer**: id, name, avatar, timestamps
- **d_album**: id, name, cover, timestamps
- **d_track**: id, name, singer, album, cover, url, lyric, duration, timestamps
- **d_user**: id, email, avatar, username, language, timestamps
- **d_media**: id, link, key, timestamps

## Implemented Models & Handlers
- **Singer**: Complete CRUD (Create, List, Update, Delete)
- **Album**: Complete CRUD
- **Track**: Complete CRUD (isolated, no relations)

## Router Endpoints
- `/singer` (GET/POST) → singer.DefaultHandler
- `/album` (GET/POST) → album.DefaultHandler
- `/track` (GET/POST) → track.DefaultHandler

## Implementation Pattern
Each model follows identical structure:
1. Model in `model/<name>.go` with TableName(), Get(), List*(), Create(), Update(), Delete()
2. Handlers in `handler/<name>/` with base.go, Create*.go, List*.go, Update*.go, Delete*.go
3. Routes registered in router.go with import and GET/POST endpoints
