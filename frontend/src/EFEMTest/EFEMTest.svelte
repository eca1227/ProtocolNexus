<script>
    import ControlPanel from './ControlPanel.svelte';

    let loadports = [
        { name: 'LP1', port: 0, run: true, stop: false, origin: true, overrun: false, alarm: false, err: 'N/A', drt: 'DRT2', org: 'ORG2', writeCmd: '' },
        { name: 'LP2', port: 0, run: false, stop: true, origin: false, overrun: false, alarm: false, err: 'N/A', drt: 'DRT2', org: 'ORG2', writeCmd: '' },
        { name: 'LP3', port: 0, run: false, stop: true, origin: false, overrun: true, alarm: false, err: 'OVERRUN', drt: 'DRT3', org: 'ORG3', writeCmd: '' },
        { name: 'LP4', port: 0, run: false, stop: true, origin: false, overrun: false, alarm: true, err: 'ALARM', drt: 'DRT4', org: 'ORG4', writeCmd: '' },
    ];
    let selectedLpIndex = 0;

    let aligners = [
        { name: 'ALG1', port: 0, run: false, stop: true, origin: false, overrun: false, alarm: false, err: 'N/A', drt: 'DRT1', org: 'ORG1', writeCmd: '' },
        { name: 'ALG2', port: 0, run: true, stop: false, origin: false, overrun: false, alarm: false, err: 'N/A', drt: 'DRT2', org: 'ORG2', writeCmd: '' },
    ];

    let selectedAlnIndex = 0;

    let wtrs = [
        { name: 'WTR1', port: 5, run: true, stop: false, origin: false, overrun: false, alarm: false, err: 'N/A', drt: 'DRT1', org: 'ORG1', log: 'WTR1 Initialized.\nReady.', writeCmd: '' },
        { name: 'WTR2', port: 0, run: false, stop: true, origin: false, overrun: false, alarm: false, err: 'N/A', drt: 'DRT2', org: 'ORG2', log: 'WTR2 Standby.', writeCmd: '' },
        { name: 'WTR3', port: 0, run: false, stop: true, origin: false, overrun: false, alarm: true, err: 'GRIP_FAIL', drt: 'DRT3', org: 'ORG3', log: 'WTR3 Alarm: Gripper Failed.', writeCmd: '' },
    ];
    let selectedWtrIndex = 0;

    function navigate(type, direction) {
        if (type === 'lp') {
            selectedLpIndex = (selectedLpIndex + direction + loadports.length) % loadports.length;
        } else if (type === 'aln') {
            selectedAlnIndex = (selectedAlnIndex + direction + aligners.length) % aligners.length;
        } else if (type === 'wtr') {
            selectedWtrIndex = (selectedWtrIndex + direction + wtrs.length) % wtrs.length;
        }
    }
</script>

<ControlPanel title="Loadport" items={loadports}>
    <div slot="actions">
        <textarea style="min-height: 100px; resize: none;" placeholder="Map Data (MLD)"></textarea>
        <div class="grid-2">
            <button class="btn">Load</button>
            <button class="btn">Unload</button>
        </div>
    </div>
</ControlPanel>

<ControlPanel title="Aligner" items={aligners}>
    <div slot="actions">
        <textarea style="min-height: 100px; resize: none;" bind:value={wtrs[0].log}></textarea>
        <div class="grid-4">
            <button class="btn">Align</button>
            <button class="btn">Reset</button>
            <button class="btn">VacOn</button>
            <button class="btn">VacOff</button>
        </div>
    </div>
</ControlPanel>

<ControlPanel title="WTR" items={wtrs}>
    <div slot="actions">
        <textarea style="min-height: 100px; resize: none;" bind:value={wtrs[0].log}></textarea>
        <div class="wtr-action-bar">
            <input type="text" placeholder="Stage" style="border: none;"/>
            <input type="text" placeholder="Slot" style="border: none;"/>
            <input type="text" placeholder="Hand" style="border: none;"/>
            <button class="btn">GET</button>
            <button class="btn">PUT</button>
        </div>
    </div>
</ControlPanel>
