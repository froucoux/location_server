function Connect(msgHandlers, locHandlers) {
	var thisConn = this;
	var handleLoc = function(loc) {
		locHandlers.forEach(function(handler) {handler.handleLoc(loc)});
	}
	var handleMsg = function(msg) {
		msg.Content = JSON.parse(msg.Content);
		msgHandlers.forEach(function(handler) {handler.handleMsg(msg)});
	}
	this.msgHandlers = msgHandlers;
	this.locHandlers = locHandlers;
	this.msgService = new WSClient("Message", "ws://178.79.176.206:8003/msg", handleMsg, function(){}, function() {});
	this.locService = new WSClient("Location", "ws://178.79.176.206:8002/loc", handleLoc, function(){}, function() {});
	this.handleMsgLocal = handleMsg;
	this.msgService.connect();
	this.locService.connect();
	this.unackedMsgs = new LinkedList();
	this.usrId = getId();
	var addMsg = new Add(this.usrId);
	this.msgService.jsonsend(addMsg);
	this.locService.jsonsend(addMsg);
	var lsvc = this.locService;
	var initLoc = function(position) {
		lat = position.coords.latitude;
		lng = position.coords.longitude;
		var locMsg = new InitLoc(lat, lng);
		lsvc.jsonsend(locMsg)
	}
	setInitCoords(initLoc);
}

Connect.prototype.sendMsg = function(to, content) {
	var msg = new Msg(to, JSON.stringify(content));
	this.msgService.jsonsend(msg);
}

Connect.prototype.sendLoc = function(loc) {
	this.locService.jsonsend(loc);
}

Connect.prototype.addMsgHandler = function(handler) {
	this.msgHandlers.append(handler);
}

Connect.prototype.rmvMsgHandler = function(handler) {
	this.msgHandlers.filter(function(l) {return handler == l;});
}

Connect.prototype.addLocHandler = function(handler) {
	this.locHandlers.append(handler);
}

Connect.prototype.rmvLocHandler = function(handler) {
	this.locHandlers.filter(function(l) {return handler == l;});
}

Connect.prototype.close = function() {
	this.msgService.close();
	this.locService.close();
}

function SyncRequest() {
	return {isSyncRequest: true};
}

function SyncResponse() {
	return {isSyncResponse: true};
}

Connect.prototype.sync = function(idMe, idYou, fun) {
	var synced = false;
	var thisConn = this;
	// NB: The correctness of this approach relies on the interval function being unable to run even once before this function has completed
	// Otherwise the SyncRequest might be sent, and responded to, before the syncHandler is registered (just echos of threading paranoia)
	var intervalId = setInterval(function() {thisConn.sendMsg(idYou, SyncRequest());}, 300);
	var syncHandler = function(msg) {
		var from = msg.From;
		var content = msg.Content;
		if (content.isSyncRequest) {
			var name = content.name;
			if (from == idYou) {
				clearInterval(intervalId);
				thisConn.rmvMsgHandler(syncHandler);
				thisConn.sendMsg(idYou, SyncResponse());
				fun();
			} else {
				console.log("Received 2sync request with unexpected id " + id + " from " + from);
			}
		} else if (content.isSyncResponse) {
			var name = content.name;
			if (from == idYou) {
				clearInterval(intervalId);
				thisConn.rmvMsgHandler(syncHandler);
				fun();
			} else {
				console.log("Received 2sync response with unexpected id " + id + " from " + from);
			}
		}
	}
	this.addMsgHandler({handleMsg: syncHandler});
}

function setInitCoords(initLoc) {
	if (navigator.geolocation) {
		navigator.geolocation.getCurrentPosition(initLoc, function(error) { console.log(JSON.stringify(error)), initLoc({"coords": {"latitude":1, "longitude":1}}) }); 
	} else {
		alert("Your browser does not support websockets");
	}
}

function init(position) {
	lat = position.coords.latitude;
	lng = position.coords.longitude;
	var locMsg = new InitLoc(lat, lng);
}
