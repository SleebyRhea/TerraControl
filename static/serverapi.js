function getElementInsideContainer(pID, chID) {
	var elm = document.getElementById(chID);
	var parent = elm ? elm.parentNode : {};
	return (parent.id && parent.id === pID) ? elm : {};
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

function sendMessage() {
	var xhttp = new XMLHttpRequest();
	data = getElementInsideContainer("send-server-msg", "server-input");

	if (data.classList.contains("c-field--success")) {
		xhttp.onreadystatechange = function() {
			if (xhttp.readyState == 4 && xhttp.status == 200) {
				console.log("Sent message: "+data.value)	
			}
		}

		xhttp.open("GET", "/api/server/say/"+data.value,true);
		xhttp.send()
	}
}

function verifyMessage(elm) {
	if (elm.value.length > 64 || elm.value.length == 0) {
		elm.classList.remove("c-field--success")
		elm.classList.add("c-field--error")
	} else {
		elm.classList.remove("c-field--error")
		elm.classList.add("c-field--success")
	}
}