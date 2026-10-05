# Test mode: classroom tests

A teacher gives a word list to a class as a test. The students take it
in Recuerdo (File > Test Mode) or in a web browser. The server checks the
answers with the same rules as practice (OpenTeacher's notation, formulas,
words with markup). The teacher looks the answers over, can count an
answer as right after all, and publishes the results. After that the
students can see them.

Everything runs from one program: `recuerdo testserver`.

## Starting a server

```sh
recuerdo testserver -admin beheerder
```

- `-admin NAME` makes the first admin account. Its password is printed once.
  Write it down, then log in with it and make the teachers' and students'
  accounts. You can add a whole class at once as "name,password" lines.
- The data is kept in one SQLite file: `testserver.db` in Recuerdo's
  configuration folder (on Linux `~/.config/recuerdo/`), or wherever
  `-db FILE` points. To make a backup, copy that file while the server is
  stopped.
- The web version is served next to the server, so students need nothing
  installed. It is found in Recuerdo's installed data. In a source tree,
  build it first with `scripts/build-web.sh`.

## Where it runs

How a school runs the server is up to the school. There are three common
ways:

1. **On one computer** (for trying it out): the default address is
   `127.0.0.1:8770`, so only that computer can use it.

2. **On the school network:**

   ```sh
   recuerdo testserver -addr :8770 -tls
   ```

   `-tls` makes the server use HTTPS with a certificate it makes for
   itself. The certificate is made once and kept next to the database.
   The server prints the certificate's fingerprint:

   ```
   Certificate fingerprint (SHA-256), to check when connecting:
     D3:51:1E:4C:...:AD:2D
   ```

   - **Desktop:** the first time Recuerdo connects to the server, it
     shows this fingerprint. Compare the two, and Recuerdo remembers the
     server. If the certificate ever changes, Recuerdo asks again.
   - **Browsers:** they warn about a certificate they don't know. They
     usually show its fingerprint too, so the check is the same.

   Students go to `https://<the computer's address>:8770/` and choose
   Test mode.

3. **Behind the school's web server** (with the school's own certificate):
   run `recuerdo testserver -addr 127.0.0.1:8770` and let the web server
   pass a path or a name to it, for example with Caddy:

   ```
   tests.school.example {
       reverse_proxy 127.0.0.1:8770
   }
   ```

   Over plain HTTP the server works too, but logins then cross the
   network unencrypted. Use HTTPS (option 2 or 3) whenever students log
   in from other computers.

## Accounts and roles

- **Admins** make accounts (admin, teacher, student) and groups (classes).
  They can see all tests.
- **Teachers** make tests of a lesson (in the web version: a lesson file
  or the open lesson; on the desktop: the open lesson). They give tests
  to students and groups, open and close them for handing in, check the
  hand-ins, count answers as right after all, and publish the results.
- **Students** see the tests given to them. They get the questions,
  never the answers. They hand in once, and see their result once it is
  published.

Anyone can change their own password. A new password ends that account's
sessions. Sessions last 12 hours. Passwords are stored as bcrypt hashes.

## The API

Recuerdo's clients use a small JSON API under `/api/`:

- `POST /api/login` gives a session token. The web version gets it as a
  cookie; other clients send it as `Authorization: Bearer ...`.
- Other calls: users, groups, tests, assignments, answers, results and
  publishing.

`internal/testserver/api.go` lists every call and who may make it. A
request with a body has to be JSON, so a form on another website cannot
use a student's login.

## OpenTeacher

OpenTeacher had a test mode that never shipped: its menu entry was
switched off. It used a small Django test server.

Recuerdo keeps its model: roles, groups, tests, hand-ins and checked
results. It fixes OpenTeacher's problems:

- checking now happens on the server;
- students can see their results;
- passwords are hashed properly;
- certificates are checked.

OpenTeacher Web, a separate online service for accounts and syncing word
lists, is not part of Recuerdo. Its lessons are kept as files.
