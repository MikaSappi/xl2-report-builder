package acoustics

import "xl2-report-builder/internal/model"

// AWeighting corrections in dB for 1/3 octave bands from 6.3 Hz to 20 kHz.
// Per IEC 61672-1.
var AWeighting = [model.BandCount]float64{
	-85.4, -77.8, -70.4, -63.4, -56.7, -50.5, -44.7, -39.4, // 6.3 - 31.5 Hz
	-34.6, -30.2, -26.2, -22.5, -19.1, -16.1, -13.4, -10.9, // 40 - 200 Hz
	-8.6, -6.6, -4.8, -3.2, -1.9, -0.8, 0.0, 0.6,           // 250 - 1250 Hz
	1.0, 1.2, 1.3, 1.2, 1.0, 0.5, -0.1, -1.1,               // 1600 - 8000 Hz
	-2.5, -4.3, -6.6, -9.3,                                   // 10000 - 20000 Hz
}

// CWeighting corrections in dB for 1/3 octave bands from 6.3 Hz to 20 kHz.
// Per IEC 61672-1.
var CWeighting = [model.BandCount]float64{
	-21.3, -17.7, -14.3, -11.2, -8.5, -6.2, -4.4, -3.0, // 6.3 - 31.5 Hz
	-2.0, -1.3, -0.8, -0.5, -0.3, -0.2, -0.1, 0.0,      // 40 - 200 Hz
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,             // 250 - 1250 Hz
	-0.1, -0.2, -0.3, -0.5, -0.8, -1.3, -2.0, -3.0,     // 1600 - 8000 Hz
	-4.4, -6.2, -8.5, -11.2,                              // 10000 - 20000 Hz
}
