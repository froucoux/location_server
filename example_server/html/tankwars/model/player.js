function Player(x, name, turretLength, initPower, minPower, maxPower, powerInc, health, keyBindings) {
	this.x = x;
	this.y = 0; // This gets set automatically by the gmae loop
	this.name = name;
	this.arc = 0;
	this.power = initPower;
	this.minPower = minPower;
	this.maxPower = maxPower;
	this.powerInc = powerInc;
	this.health = health;
	this.turretLength = turretLength;
	this.keyBindings = keyBindings;
	this.incPower = incPowerPlayer;
	this.decPower = decPowerPlayer;
	this.setClear = setClearPlayer;
	this.shouldRemove = shouldRemovePlayer;
	this.render = renderPlayer;
}

function incPowerPlayer() {
	this.power += this.powerInc;
	this.power = Math.min(this.power, this.maxPower);
}

function decPowerPlayer() {
	this.power -= this.powerInc;
	this.power = Math.max(this.power, this.minPower);
}

function setClearPlayer(ctxt, hgt) {
	var x = this.x-this.turretLength;
	var y = hgt - (this.y + this.turretLength);
	var w = this.turretLength*6; // This is a cludge value to allow for clearing power % text
	var h = this.turretLength*2;
	ctxt.clearRect(x, y, w, h);
}

function shouldRemovePlayer() {
	return false;
}

function renderPlayer(ctxt, hgt) {
	if (this.health > 0) {
		ctxt.beginPath();
		ctxt.arc(this.x, hgt-this.y, 10, 0, 2*Math.PI, true);
		ctxt.closePath();
		ctxt.fill();
		turretX = this.x+this.turretLength*Math.sin(this.arc);
		turretY = hgt-(this.y+(this.turretLength*Math.cos(this.arc)));
		ctxt.beginPath();
		ctxt.moveTo(this.x, hgt-this.y);
		ctxt.lineTo(turretX,turretY);
		ctxt.closePath();
		ctxt.stroke();
		var powerP = Math.round((this.power/maxPower)*100);
		ctxt.font = "20pt Calibri-bold";
		ctxt.fillText(powerP+"%",this.x+this.turretLength, hgt-this.y);
	}
}
