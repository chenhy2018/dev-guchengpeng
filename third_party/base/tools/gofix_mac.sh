#for mac only
find . -name '*.go' | xargs sed -i "" 's|"labix\.org/v2/mgo"|"gopkg.in/mgo.v2"|g'
find . -name '*.go' | xargs sed -i "" 's|"labix\.org/v2/mgo/bson"|"gopkg.in/mgo.v2/bson"|g'
