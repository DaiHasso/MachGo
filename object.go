package machgo

// Object is a database object able to be written to the DB.
type Object interface {
    GetTableName() string
    IsSaved() bool
    SetSaved(bool)

    PreInsertActions() error
    PostInsertActions() error
}
