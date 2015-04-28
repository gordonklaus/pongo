"use strict"

var RTCPeerConnection = window.RTCPeerConnection || window.mozRTCPeerConnection || window.webkitRTCPeerConnection;
var RTCSessionDescription = window.RTCSessionDescription || window.mozRTCSessionDescription || window.webkitRTCSessionDescription;
var RTCIceCandidate = window.RTCIceCandidate || window.mozRTCIceCandidate || window.webkitRTCIceCandidate;
navigator.GetUserMedia = navigator.GetUserMedia || navigator.mozGetUserMedia || navigator.webkitGetUserMedia;

var pc;
var isInitiator;
var bytesReceived = 0;

var ws = new WebSocket("ws://localhost:12345/ws");
ws.onerror = function(evt) { console.info("websocket error: " + evt) };
ws.onmessage = function(evt) {
	bytesReceived += evt.data.length;
	if (!pc) {
		isInitiator = evt.data === "0";
		start();
		return;
	}

	var signal = JSON.parse(evt.data);
	if (signal.sdp) {
		console.info("received sdp")
		pc.setRemoteDescription(new RTCSessionDescription(signal.sdp), function() {
			if (!isInitiator) {
				pc.createAnswer(function(answer) {
					pc.setLocalDescription(answer, function() {
						ws.send(JSON.stringify({"sdp": answer}));
						console.info("sent answer")
					}, function(err) {
						console.error("setLocalDescription: " + err);
					});
				}, function(err) {
					console.error("createAnswer: " + err);
				})
			}
		}, function(err) {
			console.error("setRemoteDescription: " + err);
		});
	} else if (signal.candidate) {
		console.info("received ICE candidate " + signal.candidate.candidate);
		pc.addIceCandidate(new RTCIceCandidate(signal.candidate));
	}
};
ws.onclose = function(evt) { console.info("websocket closed") };

function start() {
	pc = new RTCPeerConnection({iceServers: []}, {optional: [{RtpDataChannels: true}]});

	// send any ice candidates to the other peer
	pc.onicecandidate = function(evt) {
        if (!evt.candidate) {
            console.info("end of ICE candidates")
            return
        }
		ws.send(JSON.stringify({"candidate": evt.candidate}));
        console.info("sent ICE candidate " + evt.candidate.candidate);
	};

	// // once remote stream arrives, show it in the remote video element
	// pc.onaddstream = function(evt) {
	// 	remoteView.src = URL.createObjectURL(evt.stream);
	// };

	if (isInitiator) {
		var ch = pc.createDataChannel("gameState");
		ch.onopen = function(event) {
			console.info("onopen");
			ch.send('message from offerer');
		}
		ch.onmessage = function(event) {
			console.info("datachannel onmessage: " + event.data);
		}
		ch.onclose = function(event) {
			console.info("datachannel onclose: " + event);
		}
		ch.onerror = function(event) {
			console.info("datachannel onerror: " + event);
		}

		pc.createOffer(function(offer) {
			pc.setLocalDescription(offer, function() {
				ws.send(JSON.stringify({"sdp": offer}));
				console.info("sent offer");
			}, function(err) {
				console.info("setLocalDescription: " + err);
			});
		}, function(err) {
			console.info("createOffer: " + err);
		});
	} else {
		pc.ondatachannel = function(event) {
			console.info("ondatachannel");
			var ch = event.channel;
			ch.onopen = function(event) {
				ch.send('message from answerer');
			}
			ch.onmessage = function(event) {
				console.info("datachannel onmessage: " + event.data);
			}
			ch.onclose = function(event) {
				console.info("datachannel onclose: " + event);
			}
			ch.onerror = function(event) {
				console.info("datachannel onerror: " + event);
			}
		}
	}
}
