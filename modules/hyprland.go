package modules

type hyprlandConfig struct {
	Enable bool
}

func hyprland(ch chan<- Message, cfg *hyprlandConfig) {
	if !cfg.Enable {
		return
	}
}
