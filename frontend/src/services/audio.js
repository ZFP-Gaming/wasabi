function writeString(view, offset, string) {
  for (let i = 0; i < string.length; i++) {
    view.setUint8(offset + i, string.charCodeAt(i));
  }
}

function encodeWav(buffer) {
  const numChannels = buffer.numberOfChannels;
  const sampleRate = buffer.sampleRate;
  const bytesPerSample = 2;
  const blockAlign = numChannels * bytesPerSample;
  const dataLength = buffer.length * blockAlign;
  const bufferLength = 44 + dataLength;

  const arrayBuffer = new ArrayBuffer(bufferLength);
  const view = new DataView(arrayBuffer);

  writeString(view, 0, "RIFF");
  view.setUint32(4, 36 + dataLength, true);
  writeString(view, 8, "WAVE");
  writeString(view, 12, "fmt ");
  view.setUint32(16, 16, true); // PCM header size
  view.setUint16(20, 1, true); // format: PCM
  view.setUint16(22, numChannels, true);
  view.setUint32(24, sampleRate, true);
  view.setUint32(28, sampleRate * blockAlign, true);
  view.setUint16(32, blockAlign, true);
  view.setUint16(34, bytesPerSample * 8, true);
  writeString(view, 36, "data");
  view.setUint32(40, dataLength, true);

  let offset = 44;
  const channels = [];
  for (let i = 0; i < numChannels; i++) {
    channels.push(buffer.getChannelData(i));
  }

  for (let i = 0; i < buffer.length; i++) {
    for (let channel = 0; channel < numChannels; channel++) {
      let sample = channels[channel][i];
      sample = Math.max(-1, Math.min(1, sample));
      view.setInt16(offset, sample < 0 ? sample * 0x8000 : sample * 0x7fff, true);
      offset += 2;
    }
  }

  return new Blob([arrayBuffer], { type: "audio/wav" });
}

function resolveName(name) {
  const base = name.replace(/\.[^./]+$/, "");
  return `${base}.wav`;
}

export async function trimAudioFile(file, startSeconds, endSeconds, targetName) {
  const arrayBuffer = await file.arrayBuffer();
  const context = new AudioContext();
  const decoded = await context.decodeAudioData(arrayBuffer);
  await context.close();

  const safeStart = Math.max(0, Math.min(startSeconds ?? 0, decoded.duration));
  const requestedEnd = endSeconds ?? decoded.duration;
  let safeEnd = Math.min(Math.max(requestedEnd, safeStart), decoded.duration);
  if (safeEnd - safeStart <= 0.01) {
    safeEnd = decoded.duration; // fallback to archivo completo si el rango es demasiado corto
  }
  const duration = safeEnd - safeStart;

  if (duration <= 0) {
    throw new Error("El recorte es demasiado corto.");
  }

  const totalFrames = Math.ceil(duration * decoded.sampleRate);
  const offline = new OfflineAudioContext(decoded.numberOfChannels, totalFrames, decoded.sampleRate);
  const source = offline.createBufferSource();
  source.buffer = decoded;
  source.connect(offline.destination);
  source.start(0, safeStart, duration);

  const rendered = await offline.startRendering();
  const trimmedBlob = encodeWav(rendered);
  const trimmedName = resolveName(targetName || file.name);

  return new File([trimmedBlob], trimmedName, { type: "audio/wav" });
}
