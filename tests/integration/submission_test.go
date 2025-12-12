package integrationpackage

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:9090/api/v1"

func TestSubmitWorkFlow(t *testing.T) {
	resp, err := http.Get("http://localhost:9090/health")
	if err != nil {
		t.Skip("Server is not running, skipping integration test")
	}
	defer resp.Body.Close()

	assignmentID := uuid.New().String()
	student1 := uuid.New().String()
	student2 := uuid.New().String()

	work1ID := uploadWork(t, assignmentID, student1, "Unique content for integration test purpose.")
	assert.NotEmpty(t, work1ID)

	work2ID := uploadWork(t, assignmentID, student2, "Unique content for integration test purpose.")
	assert.NotEmpty(t, work2ID)

	checkReport(t, work2ID, true)
}

func uploadWork(t *testing.T, assignmentID, studentID, content string) string {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("assignment_id", assignmentID)
	writer.WriteField("student_id", studentID)

	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)

	part.Write([]byte(content))

	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"/works", body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Expected 202 Accepted, got %d. Body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			WorkID string `json:"work_id"`
		} `json:"data"`
	}
	err = json.Unmarshal(respBody, &result)
	assert.NoError(t, err)

	return result.Data.WorkID
}

func checkReport(t *testing.T, workID string, expectPlagiarism bool) {
	resp, err := http.Get(baseURL + "/works/" + workID + "/reports")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Data struct {
			IsPlagiarized bool    `json:"is_plagiarized"`
			Score         float64 `json:"similarity_score"`
		} `json:"data"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	if expectPlagiarism {
		assert.True(t, result.Data.Score > 0.9, "Expected high similarity score")
	}
}
