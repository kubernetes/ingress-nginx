package annotations

// Tests:
// No auth
// Basic
//   401
//   Realm name
//   Auth ok
//   Auth error
// Digest
//   401
//   Realm name
//   Auth ok
//   Auth error
// Check return 403 if there's an error retrieving the secret
