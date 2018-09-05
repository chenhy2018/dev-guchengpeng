import QtQuick 2.4


Rectangle {

    color: "lightsteelblue"
    border.color: "gray"

    property alias text: input.text
    property alias input: input

    TextInput {
        id: input
        anchors.fill: parent
        anchors.margins: 4
        focus: true
        Text {
            text: "input sip account to call"
            font: input.font
            color: "#aaa"
            visible: !input.text
        }
    }
}

