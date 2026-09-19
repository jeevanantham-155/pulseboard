package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lords/live-polling/backend/repository"
	"github.com/lords/live-polling/backend/services/poll"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicPollController struct {
	polls *poll.Service
	votes *poll.VoteService
}
type voteRequest struct {
	OptionID string `json:"optionId"`
}

func NewPublicPollController(polls *poll.Service, votes *poll.VoteService) *PublicPollController {
	return &PublicPollController{polls: polls, votes: votes}
}

func (controller *PublicPollController) Get(context *gin.Context) {
	pollID, err := primitive.ObjectIDFromHex(context.Param("id"))
	if err != nil {
		writeError(context, http.StatusBadRequest, "invalid poll id")
		return
	}
	record, err := controller.polls.FindByID(context.Request.Context(), pollID)
	if errors.Is(err, repository.ErrPollNotFound) {
		writeError(context, http.StatusNotFound, "poll not found")
		return
	}
	if err != nil {
		writeError(context, http.StatusInternalServerError, "unable to load poll")
		return
	}
	context.JSON(http.StatusOK, gin.H{"poll": poll.ToPublicPoll(record)})
}

func (controller *PublicPollController) Vote(context *gin.Context) {
	pollID, err := primitive.ObjectIDFromHex(context.Param("id"))
	if err != nil {
		writeError(context, http.StatusBadRequest, "invalid poll id")
		return
	}
	var request voteRequest
	if err := context.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.OptionID) == "" {
		writeError(context, http.StatusBadRequest, "optionId is required")
		return
	}
	result, err := controller.votes.Cast(context.Request.Context(), pollID, request.OptionID, voterKey(context))
	if err != nil {
		controller.writeVoteError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"poll": result})
}

func (controller *PublicPollController) writeVoteError(context *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrPollNotFound):
		writeError(context, http.StatusNotFound, "poll not found")
	case errors.Is(err, repository.ErrDuplicateVote):
		writeError(context, http.StatusConflict, "this browser has already voted")
	case errors.Is(err, poll.ErrPollInactive):
		writeError(context, http.StatusConflict, "This poll has ended")
	case errors.Is(err, poll.ErrOptionInvalid), errors.Is(err, poll.ErrVoterKeyInvalid):
		writeError(context, http.StatusBadRequest, "vote is invalid")
	default:
		writeError(context, http.StatusInternalServerError, "unable to record vote")
	}
}

func voterKey(context *gin.Context) string {
	if key := strings.TrimSpace(context.GetHeader("X-Voter-Key")); key != "" {
		return key
	}
	hash := sha256.Sum256([]byte(context.ClientIP() + "|" + context.GetHeader("User-Agent")))
	return hex.EncodeToString(hash[:])
}
