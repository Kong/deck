# Releasing a version


## Prepare

- Create a `docs/changelog-v0.5.0` branch. Substitute v0.5.0 with the version
  you are releasing.
- Write the changelog, ensure the links are all created correctly and
  make sure to write the doc with end-user in mind.
- Commit and push the branch to Github. Ensure that the Markdown is rendered
  correctly. Make changes as needed. The link on the version itself won't
  resolve correctly as the tag is not yet created.
- Once the changelog looks good, push the commit to main branch.
- Ensure you have Goreleaser and Docker installed locally.

## Release

- Tag the `HEAD` with your version. In our example, we will tag it `v0.5.0`.
- Push the tag to remote (Github).
- Run Goreleaser: `goreleaser release --rm-dist`. This will create
  a release in Github and upload all the artifacts.
- Edit the release to remove all the commit messages as the content and
  instead add a link to the changelog. Refer to older releases for reference.
- Homebrew release  
  Copy the deck.rb file to `homebrew-deck` directory.
  Make sure only version and checksum is changed and rest all is left as is.
  Commit and push for the Homebrew release.
- Docker release  
  - Run the command:
    `docker build --build-arg TAG=v0.5.0 --build-arg COMMIT=bar -t hbagdi/deck:v0.5.0 .`.
    Substitue `v0.5.0` with the version you are releasing and `bar` with the 
    Git short hash of the tag.
  - Run `docker push hbagdi/deck:v0.5.0`
  - If needed, do the following for a new release (not for a back-ported
    patch release):
    - Run `docker tag hbagdi/deck:v0.5.0 hbagdi/deck:latest`
    - Run `docker push hbagdi/deck:latest`

