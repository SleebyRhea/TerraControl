function getElementInsideContainer(pID, chID) {
	var elm = document.getElementById(chID);
	var parent = elm ? elm.parentNode : {};
	return (parent.id && parent.id === pID) ? elm : {};
}

function resetElement(e) {
	e.value = null
}
function kickPlayer(plr) {
	var xhttp = new XMLHttpRequest();

	xhttp.onreadystatechange = function() {
		if (xhttp.readyState == 4 && xhttp.status == 200) {
			console.log("Kicked player: "+plr)
			
		} else if (xhttp.status == 403) {
			console.log("Failed to kick player (not found): "+plr)
		}
	}

	xhttp.open("GET", "/api/player/kick/"+plr,true);
	xhttp.send()
};

function banPlayer(plr) {
	var xhttp = new XMLHttpRequest();

	xhttp.onreadystatechange = function() {
		if (xhttp.readyState == 4 && xhttp.status == 200) {
			console.log("Banned player: "+plr)
		} else if (xhttp.status == 403) {
			console.log("Failed to ban player (not found): "+plr)
		}
	}

	xhttp.open("GET", "/api/player/ban/"+plr,true);
	xhttp.send()
};

function setMOTD() {
	var xhttp = new XMLHttpRequest();
	data = getElementInsideContainer("send-server-motd",
		"send-server-motd-input");

	xhttp.onreadystatechange = function() {
		if (xhttp.readyState == 4 && xhttp.status == 200) {
			console.log("Setting MOTD: "+data.value)	
		}

		resetElement(data)
	}

	xhttp.open("GET", "/api/server/motd/"+data.value,true);
	xhttp.send()
}

function setPassword() {
	var xhttp = new XMLHttpRequest();
	data = getElementInsideContainer("send-server-password",
		"send-server-password-input");

	xhttp.onreadystatechange = function() {
		if (xhttp.readyState == 4 && xhttp.status == 200) {
			console.log("Setting password: "+data.value)
		}

		resetElement(data)
	}

	xhttp.open("GET", "/api/server/password/"+data.value,true);
	xhttp.send()
}

function sendMessage() {
	var xhttp = new XMLHttpRequest();
	data = getElementInsideContainer("send-server-message",
		"send-server-message-input");

	if (data.classList.contains("c-field--success")) {
		xhttp.onreadystatechange = function() {
			if (xhttp.readyState == 4 && xhttp.status == 200) {
				console.log("Sent message: "+data.value)
			}

			resetElement(data)
		}

		xhttp.open("GET", "/api/server/say/"+data.value,true);
		xhttp.send()
	}
}

function verifyMessage(elm, min, max) {
	i = getElementInsideContainer(
		"send-server-div",
		"send-server-message-button")

	if (i.classList.contains("c-button--brand")) {
		i.classList.remove("c-button--brand")
	}

	if (elm.value.length > max || elm.value.length < min) {
		elm.classList.remove("c-field--success")
		elm.classList.add("c-field--error")
		i.classList.add("c-button--error")
		if (i.classList.contains("c-button--success")) {
			i.classList.remove("c-button--success")
		}
	} else {
		elm.classList.remove("c-field--error")
		elm.classList.add("c-field--success")
		i.classList.add("c-button--success")
		if (i.classList.contains("c-button--error")) {
			i.classList.remove("c-button--error")
		}
	}
}

function startServer() {
	console.log("Starting server...")
}

function restartServer() {
	console.log("Restarting server...")
}

function stopServer() {
	console.log("Stopping server...")
}

function settleLiquids() {
	var xhttp = new XMLHttpRequest();

	xhttp.onreadystatechange = function() {
		if (xhttp.readyState == 4 && xhttp.status == 200) {
			console.log("Settling server liquids")
		} else {
			console.log("Failed to run settle command")
		}
	}

	xhttp.open("GET", "/api/server/settle", true);
	xhttp.send()
}

function serverTime(time) {
	var xhttp = new XMLHttpRequest();

	if (time) {
		xhttp.onreadystatechange = function() {
			if (xhttp.readyState == 4 && xhttp.status == 200) {
				console.log("Setting time to: "+time)
				
			} else if (xhttp.status == 403) {
				console.log("Failed to run time command")
			}
		}

		xhttp.open("GET", "/api/server/time/"+time, true);
		xhttp.send()
	} else {
		xhttp.onreadystatechange = function() {
			if (xhttp.readyState == 4 && xhttp.status == 200) {
				console.log("Getting server time: ")
				
			} else {
				console.log("Failed to run time command")
			}
		}

		xhttp.open("GET", "/api/server/time", true);
		xhttp.send()
	}
}