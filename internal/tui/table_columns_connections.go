package tui

import (
	"strconv"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// connectionsSpec is the column set of the Connections tab. The local port is
// always shown; the bind address, state and age appear when there is room.
func connectionsSpec() tableSpec {
	return tableSpec{
		fixed: []column{
			{headingDirection, directionColumnWidth, priorityMandatory, chipSidePadding, directionCell},
			{headingNew, newColumnWidth, priorityMandatory, 0, newBadgeCell},
			{headingLocal, localColumnWidth, priorityMandatory, 0, localPortCell},
			{headingBind, bindColumnWidth, priorityLocalBind, 0, localAddressCell},
			{headingRemote, remoteColumnWidth, priorityMandatory, 0, remoteCell},
			{headingState, stateColumnWidth, priorityConnectionState, 0, connectionStateCell},
			{headingAge, ageColumnWidth, priorityConnectionAge, 0, ageCell},
		},
		flexHeading:  headingProcessShort,
		flexContent:  connectionProcessCell,
		minFlexWidth: connectionMinProcessWidth,
	}
}

func directionCell(palette styles, row displayRow) cell {
	if row.connection.Direction == domain.ConnectionInbound {
		return cell{text: directionInboundTag, look: palette.chipText.Background(colorInbound), chip: true}
	}
	return cell{text: directionOutboundTag, look: palette.chipText.Background(colorOutbound), chip: true}
}

func localPortCell(palette styles, row displayRow) cell {
	return cell{text: strconv.Itoa(row.connection.LocalPort), look: palette.plain}
}

func localAddressCell(palette styles, row displayRow) cell {
	return cell{text: row.connection.LocalAddress, look: palette.dim}
}

func remoteCell(palette styles, row displayRow) cell {
	return cell{text: remoteEndpoint(row.connection), look: palette.plain}
}

func connectionStateCell(palette styles, row displayRow) cell {
	return cell{text: row.connection.State.String(), look: palette.dim}
}

func ageCell(palette styles, row displayRow) cell {
	return cell{text: formatConnectionAge(row.connectionAge), look: palette.dim}
}

func connectionProcessCell(palette styles, row displayRow) cell {
	if row.connection.ProcessName == "" {
		return cell{text: emptyValue, look: palette.dim}
	}
	return cell{text: row.connection.ProcessName, look: palette.plain}
}
