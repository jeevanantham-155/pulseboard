package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/lords/live-polling/backend/middleware"
	"github.com/lords/live-polling/backend/repository"
	"github.com/lords/live-polling/backend/services/poll"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PollController struct {
	service *poll.Service
}

type createPollRequest struct {
	Question        string   `json:"question"`
	Options         []string `json:"options"`
	DurationSeconds int      `json:"durationSeconds"`
}

func NewPollController(service *poll.Service) *PollController {
	return &PollController{service: service}
}

func (controller *PollController) Create(context *gin.Context) {
	creatorID, ok := middleware.UserID(context)
	if !ok {
		writeError(context, http.StatusUnauthorized, "authentication required")
		return
	}
	var request createPollRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	var expiresAt *time.Time
	if request.DurationSeconds != 0 {
		if request.DurationSeconds < 60 || request.DurationSeconds > 7*24*60*60 {
			writeError(context, http.StatusBadRequest, "duration must be between 60 seconds and 7 days")
			return
		}
		expires := time.Now().UTC().Add(time.Duration(request.DurationSeconds) * time.Second)
		expiresAt = &expires
	}
	pollRecord, err := controller.service.CreateWithSettings(context.Request.Context(), creatorID, request.Question, request.Options, poll.Settings{ExpiresAt: expiresAt})
	if err != nil {
		controller.writeServiceError(context, err)
		return
	}
	context.JSON(http.StatusCreated, gin.H{"poll": pollRecord})
}

func (controller *PollController) List(context *gin.Context) {
	creatorID, ok := middleware.UserID(context)
	if !ok {
		writeError(context, http.StatusUnauthorized, "authentication required")
		return
	}
	polls, err := controller.service.ListByCreator(context.Request.Context(), creatorID)
	if err != nil {
		writeError(context, http.StatusInternalServerError, "unable to load polls")
		return
	}
	context.JSON(http.StatusOK, gin.H{"polls": polls})
}

func (controller *PollController) Get(context *gin.Context) {
	creatorID, ok := middleware.UserID(context)
	if !ok {
		writeError(context, http.StatusUnauthorized, "authentication required")
		return
	}
	pollID, err := primitive.ObjectIDFromHex(context.Param("id"))
	if err != nil {
		writeError(context, http.StatusBadRequest, "invalid poll id")
		return
	}
	pollRecord, err := controller.service.FindByID(context.Request.Context(), pollID)
	if errors.Is(err, repository.ErrPollNotFound) || (err == nil && pollRecord.CreatorID != creatorID) {
		writeError(context, http.StatusNotFound, "poll not found")
		return
	}
	if err != nil {
		writeError(context, http.StatusInternalServerError, "unable to load poll")
		return
	}
	context.JSON(http.StatusOK, gin.H{"poll": pollRecord})
}

func (controller *PollController) Delete(context *gin.Context) {
	creatorID, ok := middleware.UserID(context)
	if !ok {
		writeError(context, http.StatusUnauthorized, "authentication required")
		return
	}
	pollID, err := primitive.ObjectIDFromHex(context.Param("id"))
	if err != nil {
		writeError(context, http.StatusBadRequest, "invalid poll id")
		return
	}
	if err := controller.service.Delete(context.Request.Context(), pollID, creatorID); err != nil {
		if errors.Is(err, repository.ErrPollNotFound) {
			writeError(context, http.StatusNotFound, "poll not found")
			return
		}
		writeError(context, http.StatusInternalServerError, "unable to delete poll")
		return
	}
	context.Status(http.StatusNoContent)
}

func (controller *PollController) ResultsPDF(context *gin.Context) {
	creatorID, ok := middleware.UserID(context)
	if !ok {
		writeError(context, http.StatusUnauthorized, "authentication required")
		return
	}
	pollID, err := primitive.ObjectIDFromHex(context.Param("id"))
	if err != nil {
		writeError(context, http.StatusBadRequest, "invalid poll id")
		return
	}
	record, err := controller.service.FindByID(context.Request.Context(), pollID)
	if errors.Is(err, repository.ErrPollNotFound) || (err == nil && record.CreatorID != creatorID) {
		writeError(context, http.StatusNotFound, "poll not found")
		return
	}
	if err != nil {
		writeError(context, http.StatusInternalServerError, "unable to load poll")
		return
	}
	results := poll.ToPublicPoll(record)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 12, "LIVE POLL RESULTS")
	pdf.Ln(16)
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 8, "Question: ")
	pdf.SetFont("Arial", "", 13)
	pdf.MultiCell(0, 8, record.Question, "", "", false)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 8, "Status: "+record.Status)
	pdf.Ln(7)
	pdf.Cell(0, 8, "Created: "+record.CreatedAt.Format(time.RFC3339))
	pdf.Ln(7)
	if record.ExpiresAt != nil {
		pdf.Cell(0, 8, "Closes: "+record.ExpiresAt.Format(time.RFC3339))
		pdf.Ln(7)
	}
	total := 0
	for _, item := range results.Options {
		total += item.Votes
	}
	pdf.Cell(0, 8, "Total votes: "+fmt.Sprint(total))
	pdf.Ln(14)
	for _, item := range results.Options {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 7, item.Text)
		pdf.Ln(7)
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(0, 7, fmt.Sprintf("Votes: %d   Percentage: %.1f%%", item.Votes, item.Percentage))
		pdf.Ln(10)
		pdf.SetFillColor(35, 104, 90)
		pdf.Rect(20, pdf.GetY(), 160*item.Percentage/100, 5, "F")
		pdf.Ln(12)
	}
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		writeError(context, http.StatusInternalServerError, "unable to generate PDF")
		return
	}
	context.Header("Content-Type", "application/pdf")
	context.Header("Content-Disposition", `attachment; filename="poll-results-`+pollID.Hex()+`.pdf"`)
	context.Data(http.StatusOK, "application/pdf", output.Bytes())
}

func (controller *PollController) writeServiceError(context *gin.Context, err error) {
	if errors.Is(err, poll.ErrInvalidPoll) {
		writeError(context, http.StatusBadRequest, "question and options are invalid")
		return
	}
	writeError(context, http.StatusInternalServerError, "unable to create poll")
}
