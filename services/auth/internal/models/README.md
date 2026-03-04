# models/

This directory contains data model definitions for your application, typically as Go structs.
Models here are used by GORM for ORM mapping to database tables.

## Example

```go
package models

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string
	Email string
}
```

Place your application's core data structures here.
