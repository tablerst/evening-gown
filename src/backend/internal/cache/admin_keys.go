package cache

// AdminContactsNewCountKey stores the current number of "new" contact leads.
//
// Design:
// - Strong-ish consistency is achieved by updating this counter on every write path
//   (public create + admin status change + admin delete).
// - The /admin/contacts/unread-count endpoint supports force recompute from DB.
const AdminContactsNewCountKey = "eg:admin:contacts:new-count:v1"
