package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/DSSD-Madison/gmu/db"
	"github.com/DSSD-Madison/gmu/handlers"
)

// Create a testable version of the handler function that accepts an interface
// instead of the concrete *db.Queries type
type DBQuerier interface {
	GetDocumentsByURIs(ctx context.Context, uris []string) ([]db.Document, error)
}

// MockDBQueries implements our DBQuerier interface for testing
type MockDBQueries struct {
	GetDocumentsByURIsFunc func(ctx context.Context, uris []string) ([]db.Document, error)
}

func (m *MockDBQueries) GetDocumentsByURIs(ctx context.Context, uris []string) ([]db.Document, error) {
	return m.GetDocumentsByURIsFunc(ctx, uris)
}

// HandlersTestSuite defines the test suite
type HandlersTestSuite struct {
	suite.Suite
	echo       *echo.Echo
	mockQuery  *MockDBQueries
	ctx        echo.Context
	sampleDocs []db.Document
	doc1ID     uuid.UUID
	doc2ID     uuid.UUID
	regionID1  uuid.UUID
	regionID2  uuid.UUID
	now        sql.NullTime
}

// SetupTest runs before each test
func (suite *HandlersTestSuite) SetupTest() {
	// Create a new Echo instance
	suite.echo = echo.New()

	// Create a new request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Create an Echo context
	suite.ctx = suite.echo.NewContext(req, rec)

	// Create a new mock
	suite.mockQuery = &MockDBQueries{}

	// Setup sample data
	suite.doc1ID = uuid.New()
	suite.doc2ID = uuid.New()
	suite.regionID1 = uuid.New()
	suite.regionID2 = uuid.New()
	suite.now = sql.NullTime{Time: time.Now(), Valid: true}

	// Create sample documents
	suite.sampleDocs = []db.Document{
		{
			ID:            suite.doc1ID,
			FileName:      "doc1.pdf",
			Title:         "Document 1",
			Abstract:      sql.NullString{String: "Abstract 1", Valid: true},
			Category:      sql.NullString{String: "Category 1", Valid: true},
			PublishDate:   suite.now,
			Source:        sql.NullString{String: "Source 1", Valid: true},
			RegionID:      uuid.NullUUID{UUID: suite.regionID1, Valid: true},
			S3File:        "doc1.pdf",
			S3FilePreview: sql.NullString{String: "preview1.pdf", Valid: true},
			PdfLink:       sql.NullString{String: "http://example.com/doc1.pdf", Valid: true},
			CreatedAt:     suite.now,
			DeletedAt:     sql.NullTime{Valid: false},
		},
		{
			ID:            suite.doc2ID,
			FileName:      "doc2.pdf",
			Title:         "Document 2",
			Abstract:      sql.NullString{String: "Abstract 2", Valid: true},
			Category:      sql.NullString{String: "Category 2", Valid: true},
			PublishDate:   suite.now,
			Source:        sql.NullString{String: "Source 2", Valid: true},
			RegionID:      uuid.NullUUID{UUID: suite.regionID2, Valid: true},
			S3File:        "doc2.pdf",
			S3FilePreview: sql.NullString{String: "preview2.pdf", Valid: true},
			PdfLink:       sql.NullString{String: "http://example.com/doc2.pdf", Valid: true},
			CreatedAt:     suite.now,
			DeletedAt:     sql.NullTime{Valid: false},
		},
	}
}

// TestGetDocumentsSuccessful tests the successful retrieval of documents
func (suite *HandlersTestSuite) TestGetDocumentsSuccessful() {
	// Test data
	uris := []string{"doc1.pdf", "doc2.pdf"}

	// Setup expectations
	suite.mockQuery.GetDocumentsByURIsFunc = func(ctx context.Context, u []string) ([]db.Document, error) {
		assert.Equal(suite.T(), uris, u, "URIs should match")
		return suite.sampleDocs, nil
	}

	// Expected result
	expected := map[string]db.Document{
		"doc1.pdf": suite.sampleDocs[0],
		"doc2.pdf": suite.sampleDocs[1],
	}

	// Call the function
	result, err := handlers.GetDocuments(suite.ctx, suite.mockQuery, uris)

	// Assert results
	suite.NoError(err)
	suite.Equal(expected, result)
}

// TestGetDocumentsDatabaseError tests the case when the database returns an error
func (suite *HandlersTestSuite) TestGetDocumentsDatabaseError() {
	// Test data
	uris := []string{"doc1.pdf", "doc2.pdf"}
	expectedErr := errors.New("database error")

	// Setup expectations
	suite.mockQuery.GetDocumentsByURIsFunc = func(ctx context.Context, u []string) ([]db.Document, error) {
		assert.Equal(suite.T(), uris, u, "URIs should match")
		return nil, expectedErr
	}

	// Call the function
	result, err := handlers.GetDocuments(suite.ctx, suite.mockQuery, uris)

	// Assert results
	suite.Error(err)
	suite.Equal(expectedErr.Error(), err.Error())
	suite.Nil(result)
}

// TestGetDocumentsEmptyURIs tests the case with an empty URIs list
func (suite *HandlersTestSuite) TestGetDocumentsEmptyURIs() {
	// Test data
	uris := []string{}

	// Setup expectations
	suite.mockQuery.GetDocumentsByURIsFunc = func(ctx context.Context, u []string) ([]db.Document, error) {
		assert.Equal(suite.T(), uris, u, "URIs should match")
		return []db.Document{}, nil
	}

	// Call the function
	result, err := handlers.GetDocuments(suite.ctx, suite.mockQuery, uris)

	// Assert results
	suite.NoError(err)
	suite.Equal(map[string]db.Document{}, result)
}

// TestGetDocumentsMappingLogic tests the mapping logic from slice to map
func (suite *HandlersTestSuite) TestGetDocumentsMappingLogic() {
	// Create documents with the same S3File to test overwrite behavior
	docs := []db.Document{
		{
			ID:       uuid.New(),
			FileName: "duplicate.pdf",
			Title:    "Original Document",
			S3File:   "same-key.pdf",
		},
		{
			ID:       uuid.New(),
			FileName: "another.pdf",
			Title:    "Overwriting Document",
			S3File:   "same-key.pdf",
		},
	}

	uris := []string{"same-key.pdf"}

	// Setup expectations
	suite.mockQuery.GetDocumentsByURIsFunc = func(ctx context.Context, u []string) ([]db.Document, error) {
		assert.Equal(suite.T(), uris, u, "URIs should match")
		return docs, nil
	}

	// Call the function
	result, err := handlers.GetDocuments(suite.ctx, suite.mockQuery, uris)

	// Assert results
	suite.NoError(err)
	suite.Len(result, 1)

	// The last document should win in case of duplicate keys
	suite.Equal("Overwriting Document", result["same-key.pdf"].Title)
}

// TestHandlersSuite runs the test suite
func TestHandlersSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

