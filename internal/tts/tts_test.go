package tts

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"
)

func have(progs ...string) func(string) (string, error) {
	return func(name string) (string, error) {
		for _, p := range progs {
			if p == name {
				return "/bin/" + name, nil
			}
		}
		return "", errors.New("not found")
	}
}

func TestCommand(t *testing.T) {
	name, args, _, err := Command("linux", "hond", "Dutch", have("espeak-ng"))
	if err != nil || name != "/bin/espeak-ng" || !reflect.DeepEqual(args, []string{"-v", "nl", "--", "hond"}) {
		t.Errorf("linux: %s %v %v", name, args, err)
	}
	// older espeak, a code, a text that looks like an option
	name, args, _, _ = Command("freebsd", "-dog", "EN", have("espeak"))
	if name != "/bin/espeak" || !reflect.DeepEqual(args, []string{"-v", "en", "--", "-dog"}) {
		t.Errorf("espeak: %s %v", name, args)
	}
	if _, args, _, _ := Command("linux", "x", "", have("espeak-ng")); !reflect.DeepEqual(args, []string{"--", "x"}) {
		t.Errorf("no language: %v", args)
	}
	if _, _, _, err := Command("linux", "x", "", have()); !errors.Is(err, ErrNoSpeech) {
		t.Errorf("no program: %v", err)
	}
	name, args, _, _ = Command("darwin", "dog", "en", have("say"))
	if name != "/bin/say" || !reflect.DeepEqual(args, []string{"--", "dog"}) {
		t.Errorf("mac: %s %v", name, args)
	}
	// Windows: the text goes in the environment, never into the script
	name, args, env, _ := Command("windows", `a"; rm -r x; "`, "", have("powershell"))
	if name != "/bin/powershell" || args[len(args)-1] != windowsScript || env[0] != `RECUERDO_TTS_TEXT=a"; rm -r x; "` {
		t.Errorf("windows: %s %v %v", name, args, env)
	}
}

func TestSpeaker(t *testing.T) {
	s := &Speaker{GOOS: "linux", LookPath: have()}
	if s.Available() || !errors.Is(s.Speak("x", ""), ErrNoSpeech) || s.Speak("  ", "") != nil {
		t.Error("speaker without speech")
	}
	// a stand-in speech program: "true" ignores its arguments
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("no true program here to stand in for espeak-ng")
	}
	s.LookPath = func(name string) (string, error) {
		if name == "espeak-ng" {
			return "true", nil
		}
		return "", errors.New("no")
	}
	if !s.Available() || s.Speak("hond", "nl") != nil || s.Speak("kat", "nl") != nil {
		t.Error("speaking with a stand-in failed")
	}
}
