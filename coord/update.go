package coord

import (
	"cmp"
	"slices"

	"github.com/nyiyui/qrystal/spec"
)

// updateSpec replaces Server.spec with newSpec and updates Server.latest accordingly.
// Server.specLock and Server.latestLock is taken by this function.
func (s *Server) updateSpec(newSpec spec.Spec) {
	s.specLock.Lock()
	defer s.specLock.Unlock()
	s.latestLock.Lock()
	defer s.latestLock.Unlock()
	s.updateSpecNoLock(newSpec)
}

// updateSpecNoLock replaces Server.spec and updates Server.latest accordingly.
// updateSpecNoLock does not take any locks.
// See Server.updateSpec for details.
func (s *Server) updateSpecNoLock(newSpec spec.Spec) {
	for _, oldSN := range s.spec.Networks {
		if _, ok := newSpec.GetNetwork(oldSN.Name); !ok {
			delete(s.latest, oldSN.Name)
		}
	}
	for _, newSN := range newSpec.Networks {
		oldSN, ok := s.spec.GetNetwork(newSN.Name)
		if !ok {
			continue
		}
		var keep []string
		for _, newSND := range newSN.Devices {
			_, ok := oldSN.GetDevice(newSND.Name)
			if !ok {
				continue
			}
			sndName := newSND.Name
			oldNC := oldSN.CensorForDevice(sndName)
			newNC := newSN.CensorForDevice(sndName)
			if oldNC.Equal(newNC) {
				keep = append(keep, sndName)
			}
		}
		s.latest[newSN.Name] = sliceUnion(keep, s.latest[newSN.Name])
	}
}

// sliceUnion returns the union of the two given slices.
// The slices must have unique values.
// The given slices' order may be modified.
func sliceUnion[S ~[]E, E cmp.Ordered](a, b S) S {
	slices.Sort(a)
	slices.Sort(b)
	var result S
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		x, y := a[i], b[j]
		if x == y {
			// a[1]2 3
			// b[1]2 3
			result = append(result, x)
			i++
			j++
		} else if x < y {
			// a[1]2 3
			// b[2]3
			i++
		} else if x > y {
			// a[2]3
			// b[1]2 3
			j++
		}
	}
	return result
}
