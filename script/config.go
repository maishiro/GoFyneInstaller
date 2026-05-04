package script

import (
	"io/fs"

	"gopkg.in/yaml.v3"
)

// InstallConfig はインストール仕様全体を表す
type InstallConfig struct {
	Metadata  Metadata   `yaml:"metadata"`
	Files     []FileSpec `yaml:"files"`
	Install   CommandSet `yaml:"install"`
	Uninstall CommandSet `yaml:"uninstall"`
}

// Metadata はアプリケーションメタデータ
type Metadata struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Publisher   string `yaml:"publisher"`
}

// FileSpec はバンドルするファイルの指定
type FileSpec struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
	Type   string `yaml:"type"` // "executable", "config", "data", etc.
}

// CommandSet はコマンドの集合
type CommandSet struct {
	Commands []Command `yaml:"commands"`
}

// Command はインストール/アンインストール時に実行される単一のコマンド
type Command struct {
	Type   string                 `yaml:"type"` // "copy_file", "register_service", "run_command", etc.
	Params map[string]interface{} `yaml:"params"`
}

// LoadConfigFromAssets はアセットプロバイダーから設定を読み込む
func LoadConfigFromAssets(provider AssetProvider) (*InstallConfig, error) {
	assets := provider.GetFS()
	// installer-config.yaml を読み込む（os.DirFS の場合は最初、embed.FS の場合は assets/ プリフィックス）
	data, err := fs.ReadFile(assets, "installer-config.yaml")
	if err != nil {
		// embed.FS の場合は assets/ プリフィックスで試す
		data, err = fs.ReadFile(assets, "assets/installer-config.yaml")
		if err != nil {
			return nil, err
		}
	}

	var config InstallConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// GetFileBySource はソース指定でファイルを取得
func (c *InstallConfig) GetFileBySource(source string) *FileSpec {
	for _, f := range c.Files {
		if f.Source == source {
			return &f
		}
	}
	return nil
}
