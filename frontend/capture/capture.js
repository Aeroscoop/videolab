const start = document.querySelector("#start");
const stop = document.querySelector("#stop");
const preview = document.querySelector("#preview");
const download = document.querySelector("#download");
const status = document.querySelector("#status");

let recorder = null;
let downloadURL = null;

function stopRecording() {
  if (recorder?.state === "recording") {
    stop.disabled = true;
    recorder.stop();
  }
}

start.addEventListener("click", async () => {
  start.disabled = true;
  let stream;
  try {
    stream = await navigator.mediaDevices.getDisplayMedia({
      video: {displaySurface: "browser", frameRate: {ideal: 30, max: 30}},
      audio: true,
      preferCurrentTab: false,
      selfBrowserSurface: "exclude",
      monitorTypeSurfaces: "exclude",
      systemAudio: "exclude",
    });
    const track = stream.getVideoTracks()[0];
    if (track?.getSettings().displaySurface !== "browser") {
      throw new Error("Kies een browsertabblad in het deelvenster.");
    }
    const format = ["video/webm;codecs=vp8,opus", "video/webm"]
      .find(type => MediaRecorder.isTypeSupported(type));
    if (!format) throw new Error("Deze browser kan geen WebM-opname maken.");

    const session = new MediaRecorder(stream, {mimeType: format});
    const chunks = [];
    let recordingError = false;
    session.addEventListener("dataavailable", event => {
      if (event.data.size) chunks.push(event.data);
    });
    session.addEventListener("error", () => {
      recordingError = true;
      status.textContent = "De opname is onderbroken door een browserfout.";
      stream.getTracks().forEach(track => track.stop());
    });
    session.addEventListener("stop", () => {
      stream.getTracks().forEach(track => track.stop());
      preview.srcObject = null;
      preview.hidden = true;
      recorder = null;
      stop.disabled = true;
      start.disabled = false;
      if (recordingError) return;
      if (!chunks.length) {
        status.textContent = "De opname bevat geen beeld.";
        return;
      }
      if (downloadURL) URL.revokeObjectURL(downloadURL);
      downloadURL = URL.createObjectURL(new Blob(chunks, {type: session.mimeType}));
      download.href = downloadURL;
      download.download = "videolab-tab-" + new Date().toISOString().replace(/[:.]/g, "-") + ".webm";
      download.hidden = false;
      status.textContent = "Opname klaar. Bewaar het WebM-bestand.";
    });
    track.addEventListener("ended", stopRecording, {once: true});
    session.start(1000);
    recorder = session;
    preview.srcObject = stream;
    preview.hidden = false;
    preview.play().catch(() => {});
    stop.disabled = false;
    status.textContent = stream.getAudioTracks().length ? "Opname loopt met tabbladaudio." : "Opname loopt zonder audio.";
  } catch (error) {
    stream?.getTracks().forEach(track => track.stop());
    start.disabled = false;
    status.textContent = error.name === "NotAllowedError" ? "Delen geannuleerd of geweigerd." : error.message;
  }
});
stop.addEventListener("click", stopRecording);

if (!navigator.mediaDevices?.getDisplayMedia || !window.MediaRecorder) {
  start.disabled = true;
  status.textContent = "Open deze pagina in Chrome of Chromium om een tabblad op te nemen.";
}
