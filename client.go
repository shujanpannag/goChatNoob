package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"time"

	chat "shujanpannag/goChat/protos"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type client struct {
	chat.ChatClient
	Host, Password, Name, Token string
	Shutdown                    bool
}

func Client(host, pass, name string) *client {
	return &client{
		Host:     host,
		Password: pass,
		Name:     name,
	}
}

func (c *client) Run(ctx context.Context) error {
	clientMsgChan := make(chan string, 100)

	connCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(connCtx, c.Host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "unable to connect")
	}
	defer conn.Close()

	c.ChatClient = chat.NewChatClient(conn)

	sc := bufio.NewScanner(os.Stdin)
	sc.Split(bufio.ScanLines)

	for {
		if sc.Scan() {
			msg := sc.Text()
			args := strings.Split(msg, " ")
			cmd := strings.TrimSpace(args[0])

			switch cmd {
			case "$login":
				{
					if c.Token, err = c.login(ctx); err != nil {
						return errors.WithMessage(err, "failed to login")
					}
					ClientLogf(time.Now(), "logged in successfully")
				}
			case "$logout":
				{
					ClientLogf(time.Now(), "logging out")
					if err := c.logout(ctx); err != nil {
						ClientLogf(time.Now(), "failed to log out: %v", err)
					}
				}
			case "$sin":
				{
					err = c.stream(ctx)
					if err != nil {
						return errors.WithMessage(err, "stream error")
					}
				}
			case "$msg":
				{
					if len(args) < 1 {
						ClientLogf(time.Now(), "No Messages")
					} else {
						msg := strings.Join(args, " ")
						clientMsgChan <- msg
					}
				}
			case "$exit":
				{
					os.Exit(0)
				}
			default:
				{
					ClientLogf(time.Now(), "Unknown Command")
				}
			}
		}
	}

	// if c.Token, err = c.login(ctx); err != nil {
	// 	return errors.WithMessage(err, "failed to login")
	// }
	// ClientLogf(time.Now(), "logged in successfully")

	// err = c.stream(ctx)

	// ClientLogf(time.Now(), "logging out")
	// if err := c.logout(ctx); err != nil {
	// 	ClientLogf(time.Now(), "failed to log out: %v", err)
	// }

	// return errors.WithMessage(err, "stream error")
}

func (c *client) stream(ctx context.Context) error {
	md := metadata.New(map[string]string{tokenHeader: c.Token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := c.ChatClient.Stream(ctx)
	if err != nil {
		return err
	}
	defer client.CloseSend()

	ClientLogf(time.Now(), "connected to stream")

	streamState := make(chan bool)

	go c.send(client, streamState)
	return c.receive(client, streamState)
}

func (c *client) receive(sc chat.Chat_StreamClient, ss chan bool) error {
	for {
		select {
		case <-ss:
			{
				return nil
			}
		default:
			{
				res, err := sc.Recv()

				if s, ok := status.FromError(err); ok && s.Code() == codes.Canceled {
					DebugLogf("stream canceled (usually indicates shutdown)")
					return nil
				} else if err == io.EOF {
					DebugLogf("stream closed by server")
					return nil
				} else if err != nil {
					return err
				}

				ts := tsToTime(res.Timestamp)

				switch evt := res.Event.(type) {
				case *chat.StreamResponse_ClientLogin:
					ServerLogf(ts, "%s has logged in", evt.ClientLogin.Name)
				case *chat.StreamResponse_ClientLogout:
					ServerLogf(ts, "%s has logged out", evt.ClientLogout.Name)
				case *chat.StreamResponse_ClientMessage:
					if evt.ClientMessage.Name != c.Name {
						MessageLog(ts, evt.ClientMessage.Name, evt.ClientMessage.Message)
					}
				case *chat.StreamResponse_ServerShutdown:
					ServerLogf(ts, "the server is shutting down")
					c.Shutdown = true
					return nil
				default:
					ClientLogf(ts, "unexpected event from the server: %T", evt)
					return nil
				}
			}
		}
	}
}

func (c *client) send(client chat.Chat_StreamClient, ss chan bool) {
	sc := bufio.NewScanner(os.Stdin)
	sc.Split(bufio.ScanLines)

	for {
		select {
		case <-client.Context().Done():
			DebugLogf("client send loop disconnected")
		default:
			if sc.Scan() {
				msg := sc.Text()
				args := strings.Split(msg, " ")
				cmd := strings.TrimSpace(args[0])
				if cmd == "$sout" {
					ss <- true
					return
				}

				if err := client.Send(&chat.StreamRequest{Message: msg}); err != nil {
					ClientLogf(time.Now(), "failed to send message: %v", err)
					return
				}
			} else {
				ClientLogf(time.Now(), "input scanner failure: %v", sc.Err())
				return
			}
		}
	}
}

func (c *client) login(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := c.ChatClient.Login(ctx, &chat.LoginRequest{
		Name:     c.Name,
		Password: c.Password,
	})

	if err != nil {
		return "", err
	}

	return res.Token, nil
}

func (c *client) logout(ctx context.Context) error {
	if c.Shutdown {
		DebugLogf("unable to logout (server sent shutdown signal)")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := c.ChatClient.Logout(ctx, &chat.LogoutRequest{Token: c.Token})
	if s, ok := status.FromError(err); ok && s.Code() == codes.Unavailable {
		DebugLogf("unable to logout (connection already closed)")
		return nil
	}

	return err
}
