"use strict"

var RTCPeerConnection = window.RTCPeerConnection || window.mozRTCPeerConnection || window.webkitRTCPeerConnection;
var RTCSessionDescription = window.RTCSessionDescription || window.mozRTCSessionDescription || window.webkitRTCSessionDescription;
var RTCIceCandidate = window.RTCIceCandidate || window.mozRTCIceCandidate || window.webkitRTCIceCandidate;
navigator.GetUserMedia = navigator.GetUserMedia || navigator.mozGetUserMedia || navigator.webkitGetUserMedia;

var pc;
var isInitiator;
var myName;
var bytesReceived = 0;

var ws = new WebSocket("ws://localhost:12345/ws");
ws.onerror = function(evt) { console.info("websocket error: " + evt) };
ws.onmessage = function(evt) {
	bytesReceived += evt.data.length
	if (!pc) {
		isInitiator = evt.data === "0"
		start();
		return;
	}

	var signal = JSON.parse(evt.data);
	if (signal.sdp) {
		console.info(myName + " received sdp")
		pc.setRemoteDescription(new RTCSessionDescription(signal.sdp), function() {
			if (!isInitiator) {
				pc.createAnswer(function(answer) {
					pc.setLocalDescription(answer, function() {
						ws.send(JSON.stringify({"sdp": answer}));
						console.info("sent answer")
					}, function(err) {
						console.error("setLocalDescription: " + err)
					});
				}, function(err) {
					console.error("createAnswer: " + err)
				})
			}
		}, function(err) {
			console.error("setRemoteDescription: " + err)
		});
	} else if (signal.candidate) {
		console.info(myName + " received ICE candidate: " + evt.data.length)//signal.candidate.candidate)
		pc.addIceCandidate(new RTCIceCandidate(signal.candidate));
	}
};
ws.onclose = function(evt) { console.info("websocket closed") };

function start() {
	myName = isInitiator ? "offerer" : "answerer"
	pc = new RTCPeerConnection({"iceServers": []});

	// send any ice candidates to the other peer
	pc.onicecandidate = function(evt) {
		var s = JSON.stringify({"candidate": evt.candidate})
		console.info("len: " + s.length)
		ws.send(s);
	};

	// // once remote stream arrives, show it in the remote video element
	// pc.onaddstream = function(evt) {
	// 	remoteView.src = URL.createObjectURL(evt.stream);
	// };

	if (isInitiator) {
		pc.createOffer(function(offer) {
			console.info("created offer")
			pc.setLocalDescription(offer, function() {
				ws.send(JSON.stringify({"sdp": offer}));
				console.info("sent offer")
			}, function(err) {
				console.info("setLocalDescription: " + err)
			});
		}, function(err) {
			console.info("createOffer: " + err)
		});

		var ch = pc.createDataChannel("gameState");
		ch.onopen = function(event) {
			console.info("onopen");
			ch.send('message from offerer');
		}
		ch.onmessage = function(event) {
			console.info("onmessage: " + event.data);
		}
		ch.onclose = function(event) {
			console.info("onclose: " + event);
		}
		ch.onerror = function(event) {
			console.info("onerror: " + event);
		}

		pc.ondatachannel = function(event) {
			console.info("ondatachannel");
			var ch = event.channel;
			ch.onopen = function(event) {
				ch.send('message from answerer');
			}
			ch.onmessage = function(event) {
				console.info("onmessage: " + event.data);
			}
			ch.onclose = function(event) {
				console.info("onclose: " + event);
			}
			ch.onerror = function(event) {
				console.info("onerror: " + event);
			}
		}
	}
}
