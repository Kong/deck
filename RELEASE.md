eleasing a version

## Prepare

Ensure that your `go version` is at least 1.15.x.
Create a `docs/changelog-v0.5.0` branch. Substitute v0.5.0 with the version
you are releasing.
Write the changelog, ensure the links are all created correctly and
make sure to write the doc with end-user in mind.
Commit and push the branch to Github. Ensure that the Markdown is rendered
correctly. Make changes as needed. The link on the version itself won't
resolve correctly as the tag is not yet created.
Once the changelog looks good, open a PR against the main branchpush the commit to main branch. And merge the PR.
Ensure you have Goreleaser and Docker installed locally.

## Release

- Tag the `HEAD` with your version. In our example, we will tag it `v0.5.0`.
- Push the tag to remote (Github).

As of 1.8.3, the remaining steps are automated on tag pushes to Github. They are unnecessary unless the release job fails.

- Run Goreleaser: `goreleaser release --rm-dist`. This will create
  a release in Github and upload all the artifacts.
- Edit the release to remove all the commit messages as the content and
  instead add a link to the changelog. Refer to older releases for reference.
- Homebrew release  
Clone the Kong/homebrew-deck repo
`cp <deck repo>/dist/deck.rb Formula/`. Make sure only version and checksum is changed and rest all is left as is.
Git commit, git push to master

## Docker release

As of 1.8.3, these steps are automated on tag pushes to Github. They are unnecessary unless the release job fails.

Assuming you are on the TAG commit, you need to perform the following:

```
export TAG=$(git describe --abbrev=0 --tags)
export COMMIT=$(git rev-parse --short $TAG)
docker build --build-arg TAG=$TAG --build-arg COMMIT=$COMMIT -t hbagdi/deck:$TAG .
docker push hbagdi/deck:$TAG

docker tag hbagdi/deck:$TAG kong/deck:$TAG
docker push kong/deck:$TAG


# if also the latest release (not for a back-ported patch release):
docker tag hbagdi/deck:$TAG hbagdi/deck:latest
docker push hbagdi/deck:latest

docker tag hbagdi/deck:latest kong/deck:latest
docker push kong/deck:latest
```
