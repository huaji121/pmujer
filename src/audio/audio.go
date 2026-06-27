// Package audio provides simple WAV playback via SDL3 audio streams.
package audio

import "github.com/Zyko0/go-sdl3/sdl"

// Sound holds decoded WAV data ready for playback.
type Sound struct {
	data []byte
	spec sdl.AudioSpec
}

// System manages the audio device and stream.
type System struct {
	deviceID sdl.AudioDeviceID
	stream   *sdl.AudioStream
	spec     sdl.AudioSpec
}

// NewSystem opens the default audio device and creates a stream.
func NewSystem() *System {
	spec := sdl.AudioSpec{
		Format:   sdl.AUDIO_S16,
		Channels: 2,
		Freq:     44100,
	}

	devID, err := sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK.OpenAudioDevice(&spec)
	if err != nil {
		panic("failed to open audio device: " + err.Error())
	}

	stream, err := sdl.CreateAudioStream(&spec, &spec)
	if err != nil {
		panic("failed to create audio stream: " + err.Error())
	}

	if err := devID.BindAudioStream(stream); err != nil {
		panic("failed to bind audio stream: " + err.Error())
	}

	return &System{
		deviceID: devID,
		stream:   stream,
		spec:     spec,
	}
}

// LoadWAV loads a WAV file into a Sound.
func (s *System) LoadWAV(path string) *Sound {
	data, err := sdl.LoadWAV(path, &s.spec)
	if err != nil {
		panic("failed to load WAV " + path + ": " + err.Error())
	}
	return &Sound{data: data, spec: s.spec}
}

// Play plays a sound.  Multiple calls overlap (fire-and-forget).
func (s *System) Play(sound *Sound) {
	s.stream.PutData(sound.data)
}

// Quit closes the audio device.
func (s *System) Quit() {
	s.stream.Destroy()
	s.deviceID.Close()
}
