![Lovecraft-FTP](banner.png)

<div align='center'>
	<a href='#'><img src='https://img.shields.io/badge/DEMO-Offline-teal?style=for-the-badge'></a>
	<a href='https://github.com/teamcoltra/lovecraft-ftp/blob/main/LICENSE'><img src='https://img.shields.io/badge/LICENSE-Yo-blue?style=for-the-badge'></a>
	<p>/|\(;,;)/|\</p>
</div>

<br />

---

<video src="./lovecraft-ftp.webm" controls></video>

## What is Lovecraft-FTP? 🐙

Lovecraft-FTP is a playful, virtual FTP server with a fake file system. While it might look like a serious FTP server, it’s more of a honeypot designed for silly pranks and experimentation. It supports basic FTP commands like `USER`, `PASS`, `PWD`, `CWD`, `LIST`, `RETR`, `PASV`, `PORT`, and `QUIT`, and presents a mock file system structure with amusing, randomly generated content.

This project isn’t meant for anything too serious, but you could easily modify it for your own creative pranks or honeypot scenarios.

> [!IMPORTANT]
> When accessing the virtual pictures folder there is a "porn" folder which generates NSFW movie titles. My use of this program was to see if this directory got more views than other directories (implying manual searching) vs an even spread from bots. That said, that makes this project potentially NSFW. 

## What's Needed? 🦑

Currently, Lovecraft-FTP logs all commands in JSON lines format. However, it doesn't handle log rotation. Implementing log rotation or a similar log management system would enhance the project’s reliability over time.

## Features 👾

- Basic FTP command support (`USER`, `PASS`, `PWD`, `CWD`, `LIST`, `RETR`, `PASV`, `PORT`, `QUIT`)
- Randomly generated fake file system with amusing content
- Simple, lightweight Go implementation
- Fun project for experimenting with FTP server behaviors

## Getting Started 🌀

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/teamcoltra/lovecraft-ftp.git
   cd lovecraft-ftp
   ```

2. **Build the Server:**
   ```bash
   go build -o lovecraft-ftp main.go
   ```

3. **Run the Server:**
   ```bash
   ./lovecraft-ftp
   ```

4. **Connect with an FTP Client:**
   - Host: `127.0.0.1`
   - Port: `21`
   - Username and Password: Any value

> [!NOTE]
> Hey there, I'm Travis! I'm looking for a job and I might be a good fit for your company. I'm looking to get into a support or sysadmin role, but my background is in Go, PHP (inc WordPress), and networking.

## License 😱

This project is licensed under the very permissive **"Yo" license**. Do whatever you want, as long as you remember to say "Yo."

## Support 🛸

If you enjoy this project, please star this repo to show your support! Feel free to submit issues or pull requests if you have ideas for improvements or fun new features.

---

Thanks for checking out Lovecraft-FTP! And hey, if you need a creative problem solver for your team, [let's connect!](mailto:teamcoltra@gmail.com)

