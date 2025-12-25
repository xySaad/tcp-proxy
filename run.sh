mkdir -p ~/.config/systemd/user

echo '[Service]
ExecStart=/usr/bin/chisel client http://01auth.undo.it:25565 R:25555:localhost:25566
Restart=always

[Install]
WantedBy=default.target' > ~/.config/systemd/user/01proxy.service