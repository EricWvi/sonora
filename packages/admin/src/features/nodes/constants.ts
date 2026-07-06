// Placeholder byte length sent on every createFile request. The real content
// length is unknown until the file is uploaded later, so the frontend reports a
// fixed 4 MB placeholder until upload lands. The backend accepts any size.
export const FILE_PLACEHOLDER_SIZE = 4 * 1024 * 1024;
