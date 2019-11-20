package logger

import (
	"context"
	"time"

	p9p "github.com/docker/go-p9p"
	"github.com/rs/zerolog"
)

type Logger struct {
	logger  zerolog.Logger
	session p9p.Session
}

func New(logger zerolog.Logger, session p9p.Session) *Logger {
	return &Logger{
		logger:  logger,
		session: session,
	}
}

func (l *Logger) Auth(ctx context.Context, afid p9p.Fid, uname string, aname string) (qid p9p.Qid, err error) {
	defer func(t1 time.Time) {
		l.logger.Debug().
			Str("request", "auth").
			Uint32("afid", uint32(afid)).
			Str("uname", uname).
			Str("aname", aname).
			TimeDiff("duration", time.Now(), t1).
			Dict("ret", zerolog.Dict().
				Str("qid", qid.String()).
				Err(err)).
			Msg("")
	}(time.Now())

	return l.session.Auth(ctx, afid, uname, aname)
}

func (l *Logger) Attach(ctx context.Context, fid p9p.Fid, afid p9p.Fid, uname string, aname string) (qid p9p.Qid, err error) {
	defer func(t1 time.Time) {
		l.logger.Debug().
			Str("request", "attach").
			Uint32("fid", uint32(fid)).
			Uint32("afid", uint32(afid)).
			Str("uname", uname).
			Str("aname", aname).
			TimeDiff("duration", time.Now(), t1).
			Dict("ret", zerolog.Dict().
				Str("qid", qid.String()).
				Err(err)).
			Msg("")
	}(time.Now())
	return l.session.Attach(ctx, fid, afid, uname, aname)
}

func (l *Logger) Clunk(ctx context.Context, fid p9p.Fid) (err error) {
	defer func(t1 time.Time) {
		l.logger.Debug().
			Str("request", "clunk").
			Uint32("fid", uint32(fid)).
			TimeDiff("duration", time.Now(), t1).
			Dict("ret", zerolog.Dict().
				Err(err)).
			Msg("")
	}(time.Now())
	return l.session.Clunk(ctx, fid)
}

func (l *Logger) Remove(ctx context.Context, fid p9p.Fid) (err error) {
	defer func(t1 time.Time) {
		l.logger.Debug().
			Str("request", "remove").
			Uint32("fid", uint32(fid)).
			TimeDiff("duration", time.Now(), t1).
			Dict("ret", zerolog.Dict().
				Err(err)).
			Msg("")
	}(time.Now())
	return l.session.Remove(ctx, fid)
}

func (l *Logger) Walk(ctx context.Context, fid p9p.Fid, newfid p9p.Fid, names ...string) (qids []p9p.Qid, err error) {
	defer func(t1 time.Time) {
		arr := zerolog.Arr()
		for _, qid := range qids {
			arr = arr.Str(qid.String())
		}

		l.logger.Debug().
			Str("request", "walk").
			Uint32("fid", uint32(fid)).
			Uint32("newfid", uint32(newfid)).
			Strs("names", names).
			TimeDiff("duration", time.Now(), t1).
			Dict("ret", zerolog.Dict().
				Array("qids", arr).
				Err(err)).
			Msg("")
	}(time.Now())
	return l.session.Walk(ctx, fid, newfid, names...)
}

// Read follows the semantics of io.ReaderAt.ReadAtt method except it takes
// a contxt and Fid.
func (l *Logger) Read(ctx context.Context, fid p9p.Fid, p []byte, offset int64) (n int, err error) {
	defer func(t1 time.Time) {
		l.logger.Debug().
			Str("request", "read").
			Uint32("fid", uint32(fid)).
			Int64("offset", offset).
			TimeDiff("duration", time.Now(), t1).
			Dict("ret", zerolog.Dict().
				Int("n", n).
				Err(err)).
			Msg("")
	}(time.Now())
	return l.session.Read(ctx, fid, p, offset)
}

// Write follows the semantics of io.WriterAt.WriteAt except takes a context and an Fid.
//
// If n == len(p), no error is returned.
// If n < len(p), io.ErrShortWrite will be returned.
func (l *Logger) Write(ctx context.Context, fid p9p.Fid, p []byte, offset int64) (n int, err error) {
	defer func(t1 time.Time) {
		l.logger.Debug().
			Str("request", "write").
			Uint32("fid", uint32(fid)).
			Int64("offset", offset).
			TimeDiff("duration", time.Now(), t1).
			Dict("ret", zerolog.Dict().
				Int("n", n).
				Err(err)).
			Msg("")
	}(time.Now())
	return l.session.Write(ctx, fid, p, offset)
}

func (l *Logger) Open(ctx context.Context, fid p9p.Fid, mode p9p.Flag) (qid p9p.Qid, iounit uint32, err error) {
	defer func(t1 time.Time) {
		l.logger.Debug().
			Str("request", "open").
			Uint32("fid", uint32(fid)).
			Uint8("mode", uint8(mode)).
			TimeDiff("duration", time.Now(), t1).
			Dict("ret", zerolog.Dict().
				Str("qid", qid.String()).
				Uint32("iounit", iounit).
				Err(err)).
			Msg("")
	}(time.Now())
	return l.session.Open(ctx, fid, mode)
}

func (l *Logger) Create(ctx context.Context, parent p9p.Fid, name string, perm uint32, mode p9p.Flag) (p9p.Qid, uint32, error) {
	l.logger.Debug().
		Str("request", "create").
		Msg("")
	return l.session.Create(ctx, parent, name, perm, mode)
}

func (l *Logger) Stat(ctx context.Context, fid p9p.Fid) (dir p9p.Dir, err error) {
	defer func(t1 time.Time) {
		l.logger.Debug().
			Str("request", "stat").
			Uint32("fid", uint32(fid)).
			Dict("ret", zerolog.Dict().
				Dict("dir", zerolog.Dict().
					Str("name", dir.Name).
					Str("qid", dir.Qid.String()).
					Uint64("length", dir.Length).
					Time("modTime", dir.ModTime)).
				Err(err)).
			Msg("")
	}(time.Now())
	return l.session.Stat(ctx, fid)
}

func (l *Logger) WStat(ctx context.Context, fid p9p.Fid, dir p9p.Dir) error {
	l.logger.Debug().
		Str("request", "stat").
		Msg("")
	return l.session.WStat(ctx, fid, dir)
}

// Version returns the supported version and msize of the session. This
// can be affected by negotiating or the level of support provided by the
// session implementation.
func (l *Logger) Version() (msize int, version string) {
	l.logger.Debug().
		Str("request", "stat").
		Msg("")
	return l.session.Version()
}
