package handlers

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/gin-gonic/gin"
)

type UploadHandler struct {
    uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
    // Ensure upload directory exists
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.MkdirAll(uploadDir, 0755)
    }
    return &UploadHandler{
        uploadDir: uploadDir,
    }
}

func (h *UploadHandler) HandleUpload(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }

    // Generate unique filename
    ext := filepath.Ext(file.Filename)
    filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "file", ext)
    filepath := filepath.Join(h.uploadDir, filename)

    // Save file
    if err := c.SaveUploadedFile(file, filepath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }

    // Return file URL (assuming static file serving is set up)
    fileURL := fmt.Sprintf("/uploads/%s", filename)
    c.JSON(http.StatusOK, gin.H{
        "url": fileURL,
        "filename": filename,
    })
}
