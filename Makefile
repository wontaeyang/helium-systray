icons:
	$(info #### Generating Icons ####)
	2goarray StatusPos icon < ./icon/status_pos.png > ./icon/icon_status_pos.go
	2goarray StatusPosUp icon < ./icon/status_pos_up.png > ./icon/icon_status_pos_up.go
	2goarray StatusPosDown icon < ./icon/status_pos_down.png > ./icon/icon_status_pos_down.go
	2goarray StatusErr icon < ./icon/status_err.png > ./icon/icon_status_err.go
	2goarray StatusErrUp icon < ./icon/status_err_up.png > ./icon/icon_status_err_up.go
	2goarray StatusErrDown icon < ./icon/status_err_down.png > ./icon/icon_status_err_down.go
