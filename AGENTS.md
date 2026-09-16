# Repository workflow

## Branch model

- `upstream` points to the original `Wei-Shaw/sub2api` repository.
- `origin` points to our fork, `jazelly/sub2api`.
- `main` mirrors `upstream/main`. Do not develop Noop-specific changes on `main`.
- `noop` is our maintained branch. Noop-specific changes belong on `noop`, on top of `main`.

## Standard sync workflow

This is the default workflow. Do not inspect individual upstream commits during a normal sync.

Run from the repository root with a clean working tree. Commit or temporarily stash unfinished work first; never discard it. If a temporary stash is used, restore it with `git stash pop` after the rebase.

```sh
git fetch upstream main
git switch main
git merge --ff-only upstream/main
git push origin main
git switch noop
git rebase main
git push --force-with-lease --set-upstream origin noop
```

The final push creates `origin/noop` on the first sync and safely updates it on later rebases. Keep `main` as the upstream mirror and keep all Noop work on `noop`.

If the fast-forward or rebase fails, stop and resolve the specific issue deliberately. Do not force-reset `main` or overwrite a remote branch without an explicit decision.

## Exceptional upstream review

Only when the standard sync cannot proceed, or when an explicit review of incoming upstream changes is requested, inspect the commits separately:

```sh
git fetch upstream main
git log --oneline --decorate main..upstream/main
git diff --stat main...upstream/main
```

After the review and decision, return to the standard sync workflow above.
