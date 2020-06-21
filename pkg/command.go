package misaki

type Command struct {
	Name string `toml:"name" json:"name"`
	Memo string `toml:"memo" json:"memo"`
	Programs [][]string `toml:"programs" json:"programs"`
	Output bool `toml:"output" json:"output"`
}
