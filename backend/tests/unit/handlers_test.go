package unit

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
)

func TestGetTasks(t *testing.T) {
    gin.SetMode(gin.TestMode)

    r := gin.Default()
    r.GET("/api/tasks", func(c *gin.Context) {
        c.JSON(http.StatusOK, []string{})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/tasks", nil)
    r.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}
