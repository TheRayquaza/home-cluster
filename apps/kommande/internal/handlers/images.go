package handlers

import (
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (h *Handler) ServeImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	fileID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	bucket := h.db.GridFSBucket()

	// Get content type from metadata
	var fileInfo bson.M
	_ = h.db.Collection("fs.files").FindOne(r.Context(), bson.M{"_id": fileID}).Decode(&fileInfo)

	contentType := "image/jpeg"
	if metadata, ok := fileInfo["metadata"].(bson.M); ok {
		if ct, ok := metadata["content_type"].(string); ok && ct != "" {
			contentType = ct
		}
	}

	downloadStream, err := bucket.OpenDownloadStream(r.Context(), fileID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer downloadStream.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	io.Copy(w, downloadStream)
}
