package state

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/protobuf/proto"

	pb "github.com/BranLwyd/rssdl/rssdl_proto"
)

type State struct {
	filename string

	mu sync.RWMutex // protects s
	s  *pb.State
}

func Open(filename string) (*State, error) {
	var s *pb.State
	sBytes, err := ioutil.ReadFile(filename)
	if err == nil {
		s = &pb.State{}
		if err := proto.Unmarshal(sBytes, s); err != nil {
			return nil, fmt.Errorf("could not parse state: %v", err)
		}
	} else {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("could not read state file: %v", err)
		}

		// No state file. Return an empty state.
		log.Printf("State file %q does not exist. Starting fresh", filename)
		s = &pb.State{}
	}

	state := &State{
		filename: filename,
		s:        s,
	}
	// Write immediately so we'll fail out now if the state is in an unwritable location.
	if err := state.write(); err != nil {
		return nil, fmt.Errorf("could not write state file: %v", err)
	}
	return state, nil
}

func (s *State) GetOrder(name string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fs := s.s.FeedState[name]
	if fs == nil {
		return ""
	}
	return fs.Order
}

func (s *State) SetOrder(name, order string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If s.write encounters an error, we may end up with in-memory state not matching written state.
	// But that's fine -- we'll retry writes, and in the meantime we don't want to re-download already-downloaded links.

	fs := s.s.FeedState[name]
	if fs == nil {
		if s.s.FeedState == nil {
			s.s.FeedState = map[string]*pb.State_FeedState{}
		}
		fs = &pb.State_FeedState{}
		s.s.FeedState[name] = fs
	}
	fs.Order = order
	return s.write()
}

// Assumes that s.mu is already locked. (needs at least a read-lock)
func (s *State) write() error {
	sBytes, err := proto.Marshal(s.s)
	if err != nil {
		return fmt.Errorf("could not marshal state proto: %v", err)
	}

	// Use a temporary file so that updates are atomic.
	f, err := ioutil.TempFile(filepath.Dir(s.filename), ".rssdl_state_")
	if err != nil {
		return fmt.Errorf("could not create state file: %v", err)
	}
	defer func() {
		f.Close()
		if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Could not remove %q: %v", f.Name(), err)
		}
	}()
	if _, err := f.Write(sBytes); err != nil {
		return fmt.Errorf("could not write state file: %v", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("could not close state file: %v", err)
	}
	if err := os.Rename(f.Name(), s.filename); err != nil {
		return fmt.Errorf("could not rename state file: %v", err)
	}
	return nil
}
