Uploads dynamic Docker binaries to S3 for distribution.

**Note:** only to be used by authorized users.

## Usage

    docker build -t bin-upload .
    docker run -it --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -v $HOME/.gnupg/secring.gpg:/root/.gnupg/secring.gpg:ro -v $HOME/.gnupg/pubring.gpg:/root/.gnupg/pubring.gpg:ro bin-upload
