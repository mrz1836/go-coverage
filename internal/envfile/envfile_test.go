package envfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// EnvFileSuite is the test suite for the envfile package
type EnvFileSuite struct {
	suite.Suite

	tempDir string
}

func TestEnvFileSuite(t *testing.T) {
	suite.Run(t, new(EnvFileSuite))
}

func (s *EnvFileSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
}

func (s *EnvFileSuite) writeEnvFile(name, content string) string {
	path := filepath.Join(s.tempDir, name)
	err := os.WriteFile(path, []byte(content), 0o600)
	s.Require().NoError(err)
	return path
}

func (s *EnvFileSuite) TestLoad() {
	path := s.writeEnvFile("test.env", "FOO_LOAD_TEST=bar\n")
	defer func() { _ = os.Unsetenv("FOO_LOAD_TEST") }()

	err := Load(path)
	s.Require().NoError(err)
	s.Equal("bar", os.Getenv("FOO_LOAD_TEST"))
}

func (s *EnvFileSuite) TestLoad_DoesNotOverrideExisting() {
	path := s.writeEnvFile("test.env", "EXISTING_VAR_TEST=new_value\n")
	defer func() { _ = os.Unsetenv("EXISTING_VAR_TEST") }()

	// Set an existing value
	_ = os.Setenv("EXISTING_VAR_TEST", "original_value")

	err := Load(path)
	s.Require().NoError(err)
	s.Equal("original_value", os.Getenv("EXISTING_VAR_TEST"))
}

func (s *EnvFileSuite) TestOverload() {
	path := s.writeEnvFile("test.env", "OVERLOAD_VAR_TEST=new_value\n")
	defer func() { _ = os.Unsetenv("OVERLOAD_VAR_TEST") }()

	// Set an existing value
	_ = os.Setenv("OVERLOAD_VAR_TEST", "original_value")

	err := Overload(path)
	s.Require().NoError(err)
	s.Equal("new_value", os.Getenv("OVERLOAD_VAR_TEST"))
}

func (s *EnvFileSuite) TestLoad_FileNotFound() {
	err := Load(filepath.Join(s.tempDir, "nonexistent.env"))
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to read")
}

func (s *EnvFileSuite) TestParse_EmptyLines() {
	content := `
KEY1=value1

KEY2=value2

`
	result := parse(content)
	s.Equal("value1", result["KEY1"])
	s.Equal("value2", result["KEY2"])
	s.Len(result, 2)
}

func (s *EnvFileSuite) TestParse_Comments() {
	content := `# This is a full-line comment
KEY1=value1
# Another comment
KEY2=value2`
	result := parse(content)
	s.Equal("value1", result["KEY1"])
	s.Equal("value2", result["KEY2"])
	s.Len(result, 2)
}

func (s *EnvFileSuite) TestParse_QuotedValues() {
	content := `DOUBLE_QUOTED="hello world"
SINGLE_QUOTED='hello world'
UNQUOTED=hello`
	result := parse(content)
	s.Equal("hello world", result["DOUBLE_QUOTED"])
	s.Equal("hello world", result["SINGLE_QUOTED"])
	s.Equal("hello", result["UNQUOTED"])
}

func (s *EnvFileSuite) TestParse_MalformedLines() {
	content := `GOOD=value
no_equals_sign
=no_key
ALSO_GOOD=another_value`
	result := parse(content)
	s.Equal("value", result["GOOD"])
	s.Equal("another_value", result["ALSO_GOOD"])
	s.Len(result, 2)
}

func (s *EnvFileSuite) TestParse_InlineComments() {
	content := `KEY1=value1 # this is a comment
KEY2=value2	# tab before comment
KEY3=value3#nospace`
	result := parse(content)
	s.Equal("value1", result["KEY1"])
	s.Equal("value2", result["KEY2"])
	// No space before #, so it's part of the value
	s.Equal("value3#nospace", result["KEY3"])
}

func (s *EnvFileSuite) TestParse_SpecialCharacters() {
	content := `URL=https://example.com/path?query=1&other=2
PATH_WITH_EQUALS=key=value=extra
COLON_VALUE=host:port:8080`
	result := parse(content)
	s.Equal("https://example.com/path?query=1&other=2", result["URL"])
	s.Equal("key=value=extra", result["PATH_WITH_EQUALS"])
	s.Equal("host:port:8080", result["COLON_VALUE"])
}

func (s *EnvFileSuite) TestParse_RealWorldExample() {
	content := `# Coverage Configuration
GO_COVERAGE_PROVIDER=internal
GO_COVERAGE_VERSION=v1.2.0
GO_COVERAGE_USE_LOCAL=false
GO_COVERAGE_INPUT_FILE=coverage.txt
GO_COVERAGE_OUTPUT_DIR=.
GO_COVERAGE_THRESHOLD=65.0
GO_COVERAGE_ALLOW_LABEL_OVERRIDE=true
GO_COVERAGE_EXCLUDE_PATHS=test/,vendor/,testdata/
GO_COVERAGE_BADGE_LOGO=2fas
GO_COVERAGE_REPORT_TITLE="Coverage Report"
`
	result := parse(content)
	s.Equal("internal", result["GO_COVERAGE_PROVIDER"])
	s.Equal("v1.2.0", result["GO_COVERAGE_VERSION"])
	s.Equal("false", result["GO_COVERAGE_USE_LOCAL"])
	s.Equal("coverage.txt", result["GO_COVERAGE_INPUT_FILE"])
	s.Equal(".", result["GO_COVERAGE_OUTPUT_DIR"])
	s.Equal("65.0", result["GO_COVERAGE_THRESHOLD"])
	s.Equal("true", result["GO_COVERAGE_ALLOW_LABEL_OVERRIDE"])
	s.Equal("test/,vendor/,testdata/", result["GO_COVERAGE_EXCLUDE_PATHS"])
	s.Equal("2fas", result["GO_COVERAGE_BADGE_LOGO"])
	s.Equal("Coverage Report", result["GO_COVERAGE_REPORT_TITLE"])
}

func (s *EnvFileSuite) TestProcessValue() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty value", "", ""},
		{"simple value", "hello", "hello"},
		{"double quoted", `"hello world"`, "hello world"},
		{"single quoted", `'hello world'`, "hello world"},
		{"inline comment", "value # comment", "value"},
		{"tab inline comment", "value\t# comment", "value"},
		{"no space hash", "value#notcomment", "value#notcomment"},
		{"whitespace trimmed", "  hello  ", "hello"},
		{"quoted with spaces", `"  hello  "`, "  hello  "},
		{"url value", "https://example.com", "https://example.com"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := processValue(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *EnvFileSuite) TestIntegration_LoadAndOverload() {
	basePath := s.writeEnvFile("base.env", `BASE_KEY=base_value
SHARED_KEY=base_shared
`)
	customPath := s.writeEnvFile("custom.env", `CUSTOM_KEY=custom_value
SHARED_KEY=custom_shared
`)
	defer func() {
		_ = os.Unsetenv("BASE_KEY")
		_ = os.Unsetenv("CUSTOM_KEY")
		_ = os.Unsetenv("SHARED_KEY")
	}()

	// Load base first (does not override)
	err := Load(basePath)
	s.Require().NoError(err)
	s.Equal("base_value", os.Getenv("BASE_KEY"))
	s.Equal("base_shared", os.Getenv("SHARED_KEY"))

	// Overload custom (overrides existing)
	err = Overload(customPath)
	s.Require().NoError(err)
	s.Equal("custom_value", os.Getenv("CUSTOM_KEY"))
	s.Equal("custom_shared", os.Getenv("SHARED_KEY"))
}

func (s *EnvFileSuite) TestLoadDir() {
	// Create multiple env files
	s.writeEnvFile("00-core.env", "CORE_VAR=core_value\nSHARED_VAR=core\n")
	s.writeEnvFile("10-tools.env", "TOOLS_VAR=tools_value\nSHARED_VAR=tools\n")
	s.writeEnvFile("90-project.env", "PROJECT_VAR=project_value\nSHARED_VAR=project\n")

	defer func() {
		_ = os.Unsetenv("CORE_VAR")
		_ = os.Unsetenv("TOOLS_VAR")
		_ = os.Unsetenv("PROJECT_VAR")
		_ = os.Unsetenv("SHARED_VAR")
	}()

	err := LoadDir(s.tempDir, false)
	s.Require().NoError(err)

	s.Equal("core_value", os.Getenv("CORE_VAR"))
	s.Equal("tools_value", os.Getenv("TOOLS_VAR"))
	s.Equal("project_value", os.Getenv("PROJECT_VAR"))
	// Last-wins: 90-project.env wins for SHARED_VAR
	s.Equal("project", os.Getenv("SHARED_VAR"))
}

func (s *EnvFileSuite) TestLoadDirSkipsLocalInCI() {
	s.writeEnvFile("00-core.env", "CORE_VAR_CI=core\n")
	s.writeEnvFile("99-local.env", "LOCAL_VAR_CI=local\n")

	defer func() {
		_ = os.Unsetenv("CORE_VAR_CI")
		_ = os.Unsetenv("LOCAL_VAR_CI")
	}()

	err := LoadDir(s.tempDir, true)
	s.Require().NoError(err)

	s.Equal("core", os.Getenv("CORE_VAR_CI"))
	// 99-local.env should be skipped
	s.Empty(os.Getenv("LOCAL_VAR_CI"))
}

func (s *EnvFileSuite) TestLoadDirIncludesLocalWhenNotCI() {
	s.writeEnvFile("00-core.env", "CORE_VAR_LOCAL=core\n")
	s.writeEnvFile("99-local.env", "LOCAL_VAR_LOCAL=local\n")

	defer func() {
		_ = os.Unsetenv("CORE_VAR_LOCAL")
		_ = os.Unsetenv("LOCAL_VAR_LOCAL")
	}()

	err := LoadDir(s.tempDir, false)
	s.Require().NoError(err)

	s.Equal("core", os.Getenv("CORE_VAR_LOCAL"))
	s.Equal("local", os.Getenv("LOCAL_VAR_LOCAL"))
}

func (s *EnvFileSuite) TestLoadDirEmptyDirectory() {
	emptyDir := filepath.Join(s.tempDir, "empty")
	err := os.Mkdir(emptyDir, 0o750)
	s.Require().NoError(err)

	err = LoadDir(emptyDir, false)
	s.Require().Error(err)
	s.ErrorIs(err, ErrNoEnvFiles)
}

func (s *EnvFileSuite) TestLoadDirNonexistentDirectory() {
	err := LoadDir(filepath.Join(s.tempDir, "nonexistent"), false)
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to access directory")
}

func (s *EnvFileSuite) TestLoadDirSortOrder() {
	// Create files in reverse order to verify lex sorting
	s.writeEnvFile("90-last.env", "ORDER_VAR=last\n")
	s.writeEnvFile("00-first.env", "ORDER_VAR=first\n")
	s.writeEnvFile("50-middle.env", "ORDER_VAR=middle\n")

	defer func() { _ = os.Unsetenv("ORDER_VAR") }()

	err := LoadDir(s.tempDir, false)
	s.Require().NoError(err)

	// 90-last.env loaded last, so it wins
	s.Equal("last", os.Getenv("ORDER_VAR"))
}

func (s *EnvFileSuite) TestLoadDirOnlyEnvFiles() {
	// Create .env files and non-.env files
	s.writeEnvFile("10-config.env", "ENV_FILE_VAR=loaded\n")
	err := os.WriteFile(filepath.Join(s.tempDir, "README.md"), []byte("# Readme"), 0o600)
	s.Require().NoError(err)
	err = os.WriteFile(filepath.Join(s.tempDir, "load-env.sh"), []byte("#!/bin/bash"), 0o600)
	s.Require().NoError(err)

	defer func() { _ = os.Unsetenv("ENV_FILE_VAR") }()

	loadErr := LoadDir(s.tempDir, false)
	s.Require().NoError(loadErr)

	s.Equal("loaded", os.Getenv("ENV_FILE_VAR"))
}

// Additional standalone tests using standard test functions

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "test.env")
	require.NoError(t, os.WriteFile(envFile, []byte("STANDALONE_LOAD=yes\n"), 0o600))
	defer func() { _ = os.Unsetenv("STANDALONE_LOAD") }()

	err := Load(envFile)
	require.NoError(t, err)
	assert.Equal(t, "yes", os.Getenv("STANDALONE_LOAD"))
}

func TestOverload(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "test.env")
	require.NoError(t, os.WriteFile(envFile, []byte("STANDALONE_OVERLOAD=new\n"), 0o600))
	defer func() { _ = os.Unsetenv("STANDALONE_OVERLOAD") }()

	t.Setenv("STANDALONE_OVERLOAD", "old")

	err := Overload(envFile)
	require.NoError(t, err)
	assert.Equal(t, "new", os.Getenv("STANDALONE_OVERLOAD"))
}

func TestLoadDir(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "00-a.env"), []byte("LOADDIR_KEY=a\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "10-b.env"), []byte("LOADDIR_KEY=b\n"), 0o600))
	defer func() { _ = os.Unsetenv("LOADDIR_KEY") }()

	err := LoadDir(tmpDir, false)
	require.NoError(t, err)
	assert.Equal(t, "b", os.Getenv("LOADDIR_KEY"))
}

func TestLoadDirNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notadir")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0o600))

	err := LoadDir(filePath, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotDirectory)
}

func TestParse_ExportPrefix(t *testing.T) {
	content := `export MY_VAR=hello
export OTHER_VAR=world`
	result := parse(content)
	assert.Equal(t, "hello", result["MY_VAR"])
	assert.Equal(t, "world", result["OTHER_VAR"])
}
