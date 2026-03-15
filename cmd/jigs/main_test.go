package main

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantOutputPath string
		wantFiles      []string
		wantHelp       bool
		wantVersion    bool
		wantErr        bool
	}{
		{
			name:           "single source file",
			args:           []string{".env.dist"},
			wantOutputPath: ".env",
			wantFiles:      []string{".env.dist"},
		},
		{
			name:           "multiple source files",
			args:           []string{".env.dist", ".env.dev"},
			wantOutputPath: ".env",
			wantFiles:      []string{".env.dist", ".env.dev"},
		},
		{
			name:           "no arguments",
			args:           []string{},
			wantOutputPath: ".env",
			wantFiles:      nil,
		},
		{
			name:           "short help flag",
			args:           []string{"-h"},
			wantHelp:       true,
			wantOutputPath: ".env",
		},
		{
			name:           "long help flag",
			args:           []string{"--help"},
			wantHelp:       true,
			wantOutputPath: ".env",
		},
		{
			name:           "short version flag",
			args:           []string{"-v"},
			wantVersion:    true,
			wantOutputPath: ".env",
		},
		{
			name:           "long version flag",
			args:           []string{"--version"},
			wantVersion:    true,
			wantOutputPath: ".env",
		},
		{
			name:           "short output flag",
			args:           []string{"-o", "/tmp/.env", ".env.dist"},
			wantOutputPath: "/tmp/.env",
			wantFiles:      []string{".env.dist"},
		},
		{
			name:           "long output flag",
			args:           []string{"--output", "/tmp/.env", ".env.dist"},
			wantOutputPath: "/tmp/.env",
			wantFiles:      []string{".env.dist"},
		},
		{
			name:           "output flag with equals",
			args:           []string{"--output=/tmp/.env", ".env.dist"},
			wantOutputPath: "/tmp/.env",
			wantFiles:      []string{".env.dist"},
		},
		{
			name:    "short output flag missing argument",
			args:    []string{"-o"},
			wantErr: true,
		},
		{
			name:    "long output flag missing argument",
			args:    []string{"--output"},
			wantErr: true,
		},
		{
			name:           "flags mixed with source files",
			args:           []string{".env.dist", "-o", "/tmp/.env", ".env.dev"},
			wantOutputPath: "/tmp/.env",
			wantFiles:      []string{".env.dist", ".env.dev"},
		},
		{
			name:           "help flag with source files",
			args:           []string{"-h", ".env.dist"},
			wantHelp:       true,
			wantOutputPath: ".env",
			wantFiles:      []string{".env.dist"},
		},
		{
			name:           "version flag with source files",
			args:           []string{".env.dist", "--version"},
			wantVersion:    true,
			wantOutputPath: ".env",
			wantFiles:      []string{".env.dist"},
		},
		{
			name:           "all flags combined",
			args:           []string{"-h", "-v", "-o", "out.env", ".env.dist"},
			wantHelp:       true,
			wantVersion:    true,
			wantOutputPath: "out.env",
			wantFiles:      []string{".env.dist"},
		},
		{
			name:           "output flag at the end",
			args:           []string{".env.dist", "--output=/custom/.env"},
			wantOutputPath: "/custom/.env",
			wantFiles:      []string{".env.dist"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if got.outputPath != tt.wantOutputPath {
				t.Errorf("outputPath = %q, want %q", got.outputPath, tt.wantOutputPath)
			}
			if got.showHelp != tt.wantHelp {
				t.Errorf("showHelp = %v, want %v", got.showHelp, tt.wantHelp)
			}
			if got.showVersion != tt.wantVersion {
				t.Errorf("showVersion = %v, want %v", got.showVersion, tt.wantVersion)
			}

			if len(got.sourceFiles) != len(tt.wantFiles) {
				t.Fatalf("sourceFiles length = %d, want %d", len(got.sourceFiles), len(tt.wantFiles))
			}
			for i, f := range got.sourceFiles {
				if f != tt.wantFiles[i] {
					t.Errorf("sourceFiles[%d] = %q, want %q", i, f, tt.wantFiles[i])
				}
			}
		})
	}
}
