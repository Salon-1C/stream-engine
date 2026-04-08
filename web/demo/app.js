let peerConnection = null;
let sessionURL = null;

const statusEl = document.getElementById("status");
const streamPathInput = document.getElementById("streamPath");
const videoEl = document.getElementById("video");
const startBtn = document.getElementById("startBtn");
const stopBtn = document.getElementById("stopBtn");

async function start() {
  try {
    if (peerConnection) {
      await stop();
    }

    const path = streamPathInput.value.trim();
    if (!path) {
      throw new Error("stream path is required");
    }

    setStatus("Creating viewer session...");
    await fetch("/api/viewers/connect", { method: "POST" });
    window.addEventListener("beforeunload", handleBeforeUnload);

    const cfgRes = await fetch(`/api/viewer-session?path=${encodeURIComponent(path)}`);
    if (!cfgRes.ok) {
      throw new Error("viewer session rejected");
    }
    const cfg = await cfgRes.json();

    peerConnection = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }]
    });

    peerConnection.ontrack = (event) => {
      videoEl.srcObject = event.streams[0];
    };

    peerConnection.addTransceiver("video", { direction: "recvonly" });
    peerConnection.addTransceiver("audio", { direction: "recvonly" });

    const offer = await peerConnection.createOffer();
    await peerConnection.setLocalDescription(offer);

    setStatus("Negotiating WebRTC...");
    const whepRes = await fetch(cfg.whep_url, {
      method: "POST",
      headers: {
        "Content-Type": "application/sdp",
        // Evita la pagina intersticial de ngrok en POST cross-origin
        "ngrok-skip-browser-warning": "true"
      },
      body: offer.sdp
    });
    if (!whepRes.ok) {
      throw new Error(`WHEP negotiation failed (${whepRes.status})`);
    }

    const answerSDP = await whepRes.text();
    sessionURL = whepRes.headers.get("Location");

    await peerConnection.setRemoteDescription({
      type: "answer",
      sdp: answerSDP
    });

    setStatus("Playing");
  } catch (error) {
    console.error(error);
    setStatus(`Error: ${error.message}`);
    await stop();
  }
}

async function stop() {
  if (peerConnection) {
    peerConnection.close();
    peerConnection = null;
  }

  if (sessionURL) {
    try {
      await fetch(sessionURL, { method: "DELETE" });
    } catch (error) {
      console.warn("failed to close WHEP session", error);
    }
    sessionURL = null;
  }

  await fetch("/api/viewers/disconnect", { method: "POST" });
  window.removeEventListener("beforeunload", handleBeforeUnload);
  setStatus("Stopped");
}

function handleBeforeUnload() {
  navigator.sendBeacon("/api/viewers/disconnect");
}

function setStatus(value) {
  statusEl.textContent = value;
}

startBtn.addEventListener("click", () => {
  void start();
});
stopBtn.addEventListener("click", () => {
  void stop();
});
