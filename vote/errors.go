package vote

import "fmt"

// ErrNoVote means a voter ID has not voted.
type ErrNoVote struct {
	VoterID string
}

func (e ErrNoVote) Error() string {
	return fmt.Sprintf("vote: voter id %s has no votes", e.VoterID)
}
