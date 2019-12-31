# how to publish changes

```
git checkout master

rm -rf public

hugo

echo 'istiobyexample.dev' > public/CNAME

cd public && git add --all && git commit -m "Publishing to gh-pages" && cd ..


git push origin gh-pages -f
```
