package feed

// pageSize defines how many feed records are returned per page for list endpoints.
// Single source of truth: adjust here if pagination size needs to change.
// Used by:
//   - repository.go (offset calculation, Range query)
//   - service.go (BuildPagination call)
const pageSize = 20
