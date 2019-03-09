package main

import (
	"github.com/pkg/errors"
)

type Context struct {
	usersByNick       map[string]*User
	usersByUid        map[string]*User
	serversBySid      map[string]*Server
	serversByHostname map[string]*Server
}

type Server struct {
	Hostname string
	Sid      string
}

type User struct {
	Nick       string
	Uid        string
	Encryption bool
	Server     *Server
}

func NewContext() *Context {
	return &Context{
		usersByNick:       make(map[string]*User),
		usersByUid:        make(map[string]*User),
		serversByHostname: make(map[string]*Server),
		serversBySid:      make(map[string]*Server),
	}
}

func (c *Context) AddUser(nick string, uid string, encryption bool, serverHostnameOrSid string) error {
	server, err := c.GetServer(serverHostnameOrSid)
	if err != nil {
		return err
	}

	user := &User{
		Nick:       nick,
		Uid:        uid,
		Encryption: encryption,
		Server:     server,
	}

	c.usersByNick[nick] = user
	c.usersByUid[uid] = user

	return nil
}

func (c *Context) GetUser(nickOrUid string) (user *User, err error) {
	if user, ok := c.usersByNick[nickOrUid]; ok {
		return user, nil
	}

	if user, ok := c.usersByUid[nickOrUid]; ok {
		return user, nil
	}

	return nil, errors.Errorf("Couldn't find an user called %s (Nick or UID)", nickOrUid)
}

func (c *Context) RemoveUser(nickOrUid string) {
	if _, ok := c.usersByNick[nickOrUid]; ok {
		delete(c.usersByNick, nickOrUid)
	}

	if _, ok := c.usersByUid[nickOrUid]; ok {
		delete(c.usersByUid, nickOrUid)
	}
}

func (c *Context) AddServer(hostname string, sid string) {
	server := &Server{
		Hostname: hostname,
		Sid:      sid,
	}

	c.serversByHostname[hostname] = server
	c.serversBySid[sid] = server
}

func (c *Context) GetServer(hostnameOrSid string) (user *Server, err error) {
	if server, ok := c.serversByHostname[hostnameOrSid]; ok {
		return server, nil
	}

	if server, ok := c.serversBySid[hostnameOrSid]; ok {
		return server, nil
	}

	return nil, errors.Errorf("Couldn't find a server called %s (Hostname or SID)", hostnameOrSid)
}

func (c *Context) GetServersHostnames() (hostnames []string) {
	for hostname, _ := range c.serversByHostname {
		hostnames = append(hostnames, hostname)
	}
	return
}

func (c *Context) RemoveServer(nickOrUid string) {
	if _, ok := c.usersByNick[nickOrUid]; ok {
		delete(c.usersByNick, nickOrUid)
	}

	if _, ok := c.usersByUid[nickOrUid]; ok {
		delete(c.usersByUid, nickOrUid)
	}
}
