# how to publish changes

```
git checkout master

rm -rf public

hugo

cd public && git add --all && git commit -m "Publishing to gh-pages" && cd ..

git push origin gh-pages -f
```
