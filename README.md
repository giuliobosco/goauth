# goauth

## Use

Learn how to use Google OAuth API, for authenticate GO application.  
For start the application clone the repository and move in the directory:

```
git clone git@github.com:giuliobosco/goauth.git && cd goauth
```

Then generate your OAuth token with the [Google Developer Console](https://console.developers.google.com/iam-admin/projects) and your token and secret insert in the file `creds.json` as follow:

```
{
        "cid": "hash.apps.googleusercontent.com",
        "csecret": "secret"
}
```

Then insert you URL in the environment variable of the `docker-compose.yml` file and deploy the system.

```
docker-compose up
```

## APIs

- Welcome API `http://example.com/goauth/`
- Get OAuth login URL `http://example.com/goauth/login`
- Auth API `http://example.com/goauth/v1/oauth`
