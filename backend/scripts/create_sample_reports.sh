#!/bin/bash

echo "Creating sample field service reports..."

# Create sample reports based on the PDF content
cat > /tmp/fsr_inverter31_replacement.txt << 'EOF'
Field Service Report
Date: April 9, 2024
Site: S2367 - Combined Site
Technician: Technician 1 / Technician 2
Work Order: 00549595
Work Type: Inverter Replacement

Work Performed:
- Arrived on site and checked in with ITS
- Checked in with site security and made our way to roof
- Located inverter 31 and proceeded to uninstall inverter
- Once old inverter is down, we installed the new replacement PVI14TL inverter
- Wired new inverter, set the modbus ID to the same within the previous inverter of 11 and powered the new inverter on
- Confirmed production locally and remotely with ITS upon check out, site comms are fine
- Brought old inverter back to Marlton warehouse

Failed Equipment:
- Manufacturer: Solectria
- Model: PVI14TL-480
- Serial Number: 11491604119
- Inverter ID: 31
- MBID: 11

Newly Installed Equipment:
- Manufacturer: Solectria
- Model: PVI14TL-480
- Serial Number: 11491543039
- Inverter ID: 31
- MBID: 11

Job Status: Complete
EOF

cat > /tmp/fsr_inverter40_arc_protect.txt << 'EOF'
Field Service Report
Date: March 25, 2024
Site: S2367 - Combined Site
Technician: Field Service Tech
Work Order: 00527631
Work Type: Inverter Outage (Single)

Inverter 40 Investigation:
SN: 11491624010
Error: Arc Protect

AC Voltage Measurements:
L1-N: 120v
L2-N: 123v
L3-N: 122v
L1-L2: 211v
L1-L3: 212v
L2-L3: 212v

DC Voltage Measurements:
S1: 428v
S2: 422v
S3: 436v
S4: 436v
S5: 434v
S6: 435v

- 0 volts to ground on positive and negative
- Tested all fuses, all are good
- Began testing 1 string at a time to see if inverter will produce on the individual strings
- Inverters passed the individual string tests

Called Solectria:
- Spoke with Technician 4
- Gave him serial number
- He informed this is first call in for this inverter
- Case #0440607261
- This is being logged as first occurrence. If the inverter goes down due to arc protect, Solectria will RMA arc board. If that fails to fix issue, next step is to RMA entire unit.
- As for now Inv 40 is back up and fully operational

Job Status: Inverter operational, case logged with manufacturer
EOF

cat > /tmp/fsr_wire_management.txt << 'EOF'
Field Service Report
Date: August 20, 2024
Site: S2367 - Combined Site
Technician: Technician 6
Work Order: 00549675
Work Type: Wire Management

Work Performed:
- Arrived on site, checked in with maintenance
- Moved to rooftop
- Wire way that jumps from array to array is completely disassembled and DC wiring is laying on roof
- Documented findings
- Began scanning roof to gather all missing pieces
- Was able to find all missing pieces and began laying them out
- We began snapping everything back together
- We then laid the DC wiring back in the trays and put the cover on top
- Once all was assembled we zip tied every 3 feet to keep the wire secured to each other
- After zip tying everything we clipped the zip ties and took photos of completed work
- We then checked the array itself for down wires
- We didn't find anything concerning with actual wiring under the panels themselves, made a few adjustments but overall the panel wiring is in good shape

Job Status: Complete
EOF

echo "Sample reports created in /tmp/"
ls -la /tmp/fsr_*.txt