package locserver

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"github.com/fmstephe/simpleid"
	"location_server/logutil"
	"location_server/msgutil/jsonutil"
	"location_server/msgutil/msgdef"
	"location_server/msgutil/msgwriter"
)

var iOpErr = errors.New("Illegal Message Op. Operation unrecognised or provided in illegal order.")
var idSet = simpleid.NewIdMap()

type user struct {
	id                   string
	lat, olat, lng, olng float64
	msgWriter            *msgwriter.W
}

func (usr *user) eq(oUsr *user) bool {
	return usr.id == oUsr.id
}

func newUser(ws *websocket.Conn) *user {
	return &user{msgWriter: msgwriter.New(ws)}
}

// A client request
type task struct {
	tId uint
	op  msgdef.ClientOp
	usr *user
}

// NB: The user here is a value, not a pointer
// A copy has been made to avoid race conditions with
// future user updates
func newTask(tId uint, op msgdef.ClientOp, usr user) *task {
	return &task{tId: tId, op: op, usr: &usr}
}

//  Listen to ws
//  Unmarshall json objects from ws and write to readChan
func WebsocketUser(ws *websocket.Conn) {
	var tId uint
	usr := newUser(ws)
	idMsg := &msgdef.CIdMsg{}
	if err := jsonutil.JSONCodec.Receive(ws, idMsg); err != nil {
		usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
		return
	}
	if err := idMsg.Validate(); err != nil {
		usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
		return
	}
	if err := processReg(idMsg, usr); err != nil {
		usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
		return
	}
	if err := idSet.Add(usr.id, usr); err != nil {
		usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
		return
	}
	logutil.Registered(tId, usr.id)
	defer removeId(&tId, usr)
	tId++
	initLocMsg := msgdef.EmptyCLocMsg()
	if err := jsonutil.JSONCodec.Receive(ws, initLocMsg); err != nil {
		usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
		return
	}
	if err := initLocMsg.Validate(); err != nil {
		usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
		return
	}
	if err := processInitLoc(tId, initLocMsg, usr); err != nil {
		usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
		return
	}
	defer removeFromTree(&tId, usr)
	for {
		tId++
		locMsg := msgdef.EmptyCLocMsg()
		if err := jsonutil.JSONCodec.Receive(ws, locMsg); err != nil {
			usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
			return
		}
		if err := locMsg.Validate(); err != nil {
			usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
			return
		}
		if err := processRequest(tId, locMsg, usr); err != nil {
			usr.msgWriter.ErrorAndClose(tId, usr.id, err.Error())
			return
		}
	}
}

func removeId(tId *uint, usr *user) {
	(*tId)++
	logutil.Deregistered(*tId, usr.id)
	idSet.Remove(usr.id)
}

func removeFromTree(tId *uint, usr *user) {
	(*tId)++
	msg := newTask(*tId, msgdef.CRemoveOp, *usr)
	forwardMsg(msg)
}

// Handle registration message
// Success will leave usr with initialised Id field
func processReg(idMsg *msgdef.CIdMsg, usr *user) error {
	switch idMsg.Op {
	case msgdef.CAddOp:
		usr.id = idMsg.Id
		return nil
	}
	return iOpErr
}

// Handle initial location message
func processInitLoc(tId uint, initMsg *msgdef.CLocMsg, usr *user) error {
	switch initMsg.Op {
	case msgdef.CInitLocOp:
		usr.olat = initMsg.Lat
		usr.olng = initMsg.Lng
		usr.lat = initMsg.Lat
		usr.lng = initMsg.Lng
		msg := newTask(tId, msgdef.CInitLocOp, *usr)
		forwardMsg(msg)
		return nil
	}
	return iOpErr
}

// Handle request messages - cMove, cNearby
func processRequest(tId uint, locMsg *msgdef.CLocMsg, usr *user) error {
	switch locMsg.Op {
	case msgdef.CNearbyOp:
		msg := newTask(tId, msgdef.CNearbyOp, *usr)
		forwardMsg(msg)
		return nil
	case msgdef.CMoveOp:
		usr.olat = usr.lat
		usr.olng = usr.lng
		usr.lat = locMsg.Lat
		usr.lng = locMsg.Lng
		msg := newTask(tId, msgdef.CMoveOp, *usr)
		forwardMsg(msg)
		return nil
	}
	return iOpErr
}

func forwardMsg(msg *task) {
	msgChan <- msg
}
