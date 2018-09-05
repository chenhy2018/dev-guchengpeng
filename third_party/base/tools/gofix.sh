find . -name '*.go' | xargs sed 's|"labix\.org/v2/mgo"|"gopkg.in/mgo.v2"|g' -i
find . -name '*.go' | xargs sed 's|"labix\.org/v2/mgo/bson"|"gopkg.in/mgo.v2/bson"|g' -i

