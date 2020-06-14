'use strict'

function DOMLoaded() {
	return false
}

function resetElement(e) {
	e.value = null
}

function getElementInsideContainer(pID, chID) {
	var elm = document.getElementById(chID);
	var parent = elm ? elm.parentNode : {};
	return (parent.id && parent.id === pID) ? elm : {};
}

var DEBUG = true
var APIBASE = "/api/"
var APIRE = '\\/api\\/'

var playerKick     = DOMLoaded
var playerBan      = DOMLoaded
var serverSay      = DOMLoaded
var serverStop     = DOMLoaded
var serverMOTD     = DOMLoaded
var serverTime     = DOMLoaded
var serverStart    = DOMLoaded
var serverSettle   = DOMLoaded
var serverPassword = DOMLoaded
var serverRestart  = DOMLoaded
var verifyMessage  = DOMLoaded
var getRequester   = DOMLoaded

var scopes = new Map()

class TerraControlAPI {
	constructor(scope, obj) {
		TerraControlAPI.RegisterEndpoint(scope, obj, this)
		this.scope = scope
		this.obj = obj
	}

	static RequestBuilder(s, o, ...args) {
		var r = APIBASE + s + "/" + o + "/";
		if (args.length > 0) {
			for (var v of Array.from(args)) {
				if (v != "" && v != undefined) {
					r = r + v;
				}
			}
		}
		return r;
	}

	static RegisterEndpoint(s, o, n) {
		var scope = scopes.get(s);
		if ( scope == undefined) {
			console.log("TerraControlAPI: RegisterEndpoint: Invalid scope: "+s);
			return false;
		} else {
			scope.set(o, n)
		}
	}

	static Requester(r) {
		var re = new RegExp(APIRE+"[^\\/]+\\/[^\\/]+\\/")
		var s = r.match(re, "g")[0].split("/")
		return scopes.get(s[2]).get(s[3])
	}

	getdata() {
		return false
	}

	onprecall() {
		// Block all calls until the DOM is loaded or this function is overridden
		return DOMLoaded()
	}

	oncomplete() {
		if (DEBUG) {
			console.log("TerraControl API: Unimplemented: oncomplete: "+this.request)
		}
	}

	onsuccess() {
		if (DEBUG) {
			console.log("TerraControl API: Unimplemented: onsuccess: "+this.request)
		}
	}

	onredirect() {
		if (DEBUG) {
			console.log("TerraControl API: Unimplemented: onredirect: "+this.request)
		}
	}

	onfailure() {
		if (DEBUG) {
			console.log("TerraControl API: Unimplemented: onfail: "+this.request)
		}
	}

	onbadrequest() {
		if (DEBUG) {
			console.log("TerraControl API: Unimplemented: onservererror: "+this.request)
		}
	}

	call(data) {
		if (this.request) {
			this.lastrequest = this.request
		}

		this.request = TerraControlAPI.RequestBuilder(this.scope, this.obj,
			this.getdata(), data);

		if (DEBUG) {
			console.log("Making request: "+this.request)
		}
		var xhttp = new XMLHttpRequest();

		if (this.onprecall) {
			xhttp.onreadystatechange = function() {
				if (xhttp.readyState == 4) {
					var r = TerraControlAPI.Requester(this.responseURL)
					r.oncomplete(this.status)
					switch (true) {
						case is2XX(this.status):
							r.onsuccess(this.status);
							break;
						case is3XX(this.status):
							r.onredirect(this.status);
							break;
						case is4XX(this.status):
							r.onfailure(this.status);
							break;
						case is5XX(this.status):
							r.onservererror(this.status);
							break;
						default:
							console.log("TerraControl API: Invalid Response: "+this.status)
					}
				}
			}
		}

		xhttp.open("GET", this.request, true)
		xhttp.send()
	}
}

var scopes = new Map()

function is2XX(r) {
	return (r <= 299 && r >= 200)
}

function is3XX(r) {
	return (r <= 399 && r >= 300)
}

function is4XX(r) {
	return (r <= 499 && r >= 400)
}

function is5XX(r) {
	return (r <= 599 && r >= 500)
}

// BEGIN
document.addEventListener('DOMContentLoaded', () => {
	// Only permit the creation of endpoints once the DOM is loaded
	scopes.set("server", new Map())
	scopes.set("player", new Map())

	playerKick     = new TerraControlAPI("player", "kick")
	playerBan      = new TerraControlAPI("player", "ban")
	serverSay      = new TerraControlAPI("server", "say")
	serverStop     = new TerraControlAPI("server", "stop")
	serverMOTD     = new TerraControlAPI("server", "motd")
	serverTime     = new TerraControlAPI("server", "time")
	serverStart    = new TerraControlAPI("server", "start")
	serverSettle   = new TerraControlAPI("server", "settle")
	serverPassword = new TerraControlAPI("server", "password")

	// serverSay
	serverSay.onprecall = function() {
		d = getElementInsideContainer("send-server-message",
			"send-server-message-input");
		if (d.contains("c-field--success")) {
			return true
		} else {
			return false
		}
	}

	serverSay.getdata = function() {
		return getElementInsideContainer("send-server-message",
			"send-server-message-input").value;
	}

	serverSay.onsuccess = function() {
		var d = getElementInsideContainer("send-server-message",
			"send-server-message-input");
		resetElement(d)
	}


	// serverMOTD
	serverMOTD.getdata = function() {
		return getElementInsideContainer("send-server-motd",
		"send-server-motd-input").value;
	}

	serverMOTD.onsuccess = function() {
		var d = getElementInsideContainer("send-server-motd",
			"send-server-motd-input");
		resetElement(d); 
	}


	// serverPassword
	serverPassword.getdata = function() {
		return getElementInsideContainer("send-server-password",
		"send-server-password-input").value;
	}

	serverPassword.onsuccess = function() {
		var d = getElementInsideContainer("send-server-password",
			"send-server-password-input");
		resetElement(d); 
	}
	
	
	verifyMessage = function (elm, min, max) {
		var i = getElementInsideContainer("send-server-div",
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

	serverRestart = function() {
		serverStop.call()
		serverStart.call()
	}

	DOMLoaded = function() {
		return true
	}

	console.log("DOM is ready, and javascript is loaded.")
})