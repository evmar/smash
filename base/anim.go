package base

import "time"

type Anim interface {
	Frame(t time.Time) bool
}

type Lerp struct {
	val        *int
	init, targ int
	start, end time.Time
	Done       bool
}

func NewLerp(val *int, targ int, dur time.Duration) *Lerp {
	now := time.Now()
	return &Lerp{
		val:   val,
		init:  *val,
		targ:  targ,
		start: now,
		end:   now.Add(dur),
	}
}

func (l *Lerp) Frame(t time.Time) bool {
	if !t.Before(l.end) {
		*l.val = l.targ
		l.Done = true
		return false
	}
	frac := float64(t.Sub(l.start)) / float64(l.end.Sub(l.start))
	*l.val = l.init + int(frac*float64(l.targ-l.init))
	return true
}

type AnimSet struct {
	anims     map[Anim]bool
	period    time.Duration
	lastFrame time.Time
}

func NewAnimSet() *AnimSet {
	return &AnimSet{
		anims:  map[Anim]bool{},
		period: time.Duration(17 * time.Millisecond),
	}
}

func (as *AnimSet) Add(anim Anim) {
	as.anims[anim] = true
}

func (as *AnimSet) Drop(anim Anim) {
	delete(as.anims, anim)
}

func (as *AnimSet) NextFrame(force bool) <-chan time.Time {
	if !force && len(as.anims) == 0 {
		return nil
	}
	dt := time.Now().Sub(as.lastFrame)
	if dt < as.period {
		dt = as.period - dt
	} else if force {
		dt = time.Duration(0)
	}
	return time.After(dt)
}

func (as *AnimSet) Run() {
	now := time.Now()
	for anim := range as.anims {
		if !anim.Frame(now) {
			delete(as.anims, anim)
		}
	}
	as.lastFrame = now
}
