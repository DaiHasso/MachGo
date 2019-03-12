package database

import (
	"errors"
	"regexp"

	"github.com/daihasso/slogging"
)

// ErrObjectNotSaved is an error that occurs when a object is attempted to be
// updated but hasn't been saved yet.
var ErrObjectNotSaved = errors.New(
	"object cannot be updated without first being saved",
)

// ErrDuplicateEntry is an error that occurs when a duplicate is
// saved to the DB.
var ErrDuplicateEntry = errors.New("duplicate entry found in database")

// ErrNoResults is returned when you can't find a result for a
// search in the database.
var ErrNoResults = errors.New("no results found in database")

var duplicateEntryRegexPostgres = `^pq: duplicate key value`
var duplicateEntryRegexMysql = `1062:?\s*[dD]uplicate\s+(?:entry )?('[^']+')` +
	`?\s+(?:(?:(?:for key\s+)('[^']+'))|(?:[eE]ntry))`
var noResultsRegex = `.*: no rows in result set.*`

func translateDBError(err error) error {
	errorString := err.Error()

	logging.Debug("Parsing error from DB.", logging.Extras{
		"error_string": errorString,
    })

	if matched, _ := regexp.Match(
		duplicateEntryRegexMysql,
		[]byte(errorString),
	); matched {
		return ErrDuplicateEntry
	}
	if matched, _ := regexp.Match(
		duplicateEntryRegexPostgres,
		[]byte(errorString),
	); matched {
		return ErrDuplicateEntry
	}
	if matched, _ := regexp.Match(
		noResultsRegex,
		[]byte(errorString),
	); matched {
		return ErrNoResults
	}

	return err
}
