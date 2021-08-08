package main

import (
	"fmt"
	"math/rand"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func (s *server) genToken() string {
	tkn := make([]byte, 4)
	rand.Read(tkn)
	return fmt.Sprintf("%x", tkn)
}

func (s *server) getName(tkn string) (name string, ok bool) {
	s.namesMtx.RLock()
	name, ok = s.ClientNames[tkn]
	s.namesMtx.RUnlock()
	return
}

func (s *server) setName(tkn string, name string) {
	s.namesMtx.Lock()
	s.ClientNames[tkn] = name
	s.namesMtx.Unlock()
}

func (s *server) delName(tkn string) (name string, ok bool) {
	name, ok = s.getName(tkn)

	if ok {
		s.namesMtx.Lock()
		delete(s.ClientNames, tkn)
		s.namesMtx.Unlock()
	}

	return
}

func (s *server) extractToken(ctx context.Context) (tkn string, ok bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md[tokenHeader]) == 0 {
		return "", false
	}

	return md[tokenHeader][0], true
}
