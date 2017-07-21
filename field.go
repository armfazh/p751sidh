package cln16sidh

//------------------------------------------------------------------------------
// Extension Field
//------------------------------------------------------------------------------

// Represents an element of the extension field F_{p^2}.
type ExtensionFieldElement struct {
	// This field element is in Montgomery form, so that the value `a` is
	// represented by `aR mod p`.
	a fp751Element
	// This field element is in Montgomery form, so that the value `b` is
	// represented by `bR mod p`.
	b fp751Element
}

// Set dest = lhs * rhs.
//
// Allowed to overlap lhs or rhs with dest.
//
// Returns dest to allow chaining operations.
func (dest *ExtensionFieldElement) Mul(lhs, rhs *ExtensionFieldElement) *ExtensionFieldElement {
	// Let (a,b,c,d) = (lhs.a,lhs.b,rhs.a,rhs.b).
	a := &lhs.a
	b := &lhs.b
	c := &rhs.a
	d := &rhs.b

	// We want to compute
	//
	// (a + bi)*(c + di) = (a*c - b*d) + (a*d + b*c)i
	//
	// Use Karatsuba's trick: note that
	//
	// (b - a)*(c - d) = (b*c + a*d) - a*c - b*d
	//
	// so (a*d + b*c) = (b-a)*(c-d) + a*c + b*d.

	var ac, bd fp751X2
	fp751Mul(&ac, a, c) // = a*c*R*R
	fp751Mul(&bd, b, d) // = b*d*R*R

	var b_minus_a, c_minus_d fp751Element
	fp751SubReduced(&b_minus_a, b, a) // = (b-a)*R
	fp751SubReduced(&c_minus_d, c, d) // = (c-d)*R

	var ad_plus_bc fp751X2
	fp751Mul(&ad_plus_bc, &b_minus_a, &c_minus_d) // = (b-a)*(c-d)*R*R
	fp751X2AddLazy(&ad_plus_bc, &ad_plus_bc, &ac) // = ((b-a)*(c-d) + a*c)*R*R
	fp751X2AddLazy(&ad_plus_bc, &ad_plus_bc, &bd) // = ((b-a)*(c-d) + a*c + b*d)*R*R

	fp751MontgomeryReduce(&dest.b, &ad_plus_bc) // = (a*d + b*c)*R mod p

	var ac_minus_bd fp751X2
	fp751X2SubLazy(&ac_minus_bd, &ac, &bd)       // = (a*c - b*d)*R*R
	fp751MontgomeryReduce(&dest.a, &ac_minus_bd) // = (a*c - b*d)*R mod p

	return dest
}

// Set dest = 1/x
//
// Allowed to overlap dest with x.
//
// Returns dest to allow chaining operations.
func (dest *ExtensionFieldElement) Inv(x *ExtensionFieldElement) *ExtensionFieldElement {
	a := &x.a
	b := &x.b

	// We want to compute
	//
	//    1          1     (a - bi)	    (a - bi)
	// -------- = -------- -------- = -----------
	// (a + bi)   (a + bi) (a - bi)   (a^2 + b^2)
	//
	// Letting c = 1/(a^2 + b^2), this is
	//
	// 1/(a+bi) = a*c - b*ci.

	var asq_plus_bsq PrimeFieldElement
	var asq, bsq fp751X2
	fp751Mul(&asq, a, a)                         // = a*a*R*R
	fp751Mul(&bsq, b, b)                         // = b*b*R*R
	fp751X2AddLazy(&asq, &asq, &bsq)             // = (a^2 + b^2)*R*R
	fp751MontgomeryReduce(&asq_plus_bsq.a, &asq) // = (a^2 + b^2)*R mod p
	// Now asq_plus_bsq = a^2 + b^2

	var asq_plus_bsq_inv PrimeFieldElement
	asq_plus_bsq_inv.Inv(&asq_plus_bsq)
	c := &asq_plus_bsq_inv.a

	var ac fp751X2
	fp751Mul(&ac, a, c)
	fp751MontgomeryReduce(&dest.a, &ac)

	var minus_b fp751Element
	fp751SubReduced(&minus_b, &minus_b, b)
	var minus_bc fp751X2
	fp751Mul(&minus_bc, &minus_b, c)
	fp751MontgomeryReduce(&dest.b, &minus_bc)

	return dest
}

// Set dest = x * x
//
// Allowed to overlap dest with x.
//
// Returns dest to allow chaining operations.
func (dest *ExtensionFieldElement) Square(x *ExtensionFieldElement) *ExtensionFieldElement {
	a := &x.a
	b := &x.b

	// We want to compute
	//
	// (a + bi)*(a + bi) = (a^2 - b^2) + 2abi.

	var a2, a_plus_b, a_minus_b fp751Element
	fp751AddReduced(&a2, a, a)        // = a*R + a*R = 2*a*R
	fp751AddReduced(&a_plus_b, a, b)  // = a*R + b*R = (a+b)*R
	fp751SubReduced(&a_minus_b, a, b) // = a*R - b*R = (a-b)*R

	var asq_minus_bsq, ab2 fp751X2
	fp751Mul(&asq_minus_bsq, &a_plus_b, &a_minus_b) // = (a+b)*(a-b)*R*R = (a^2 - b^2)*R*R
	fp751Mul(&ab2, &a2, b)                          // = 2*a*b*R*R

	fp751MontgomeryReduce(&dest.a, &asq_minus_bsq) // = (a^2 - b^2)*R mod p
	fp751MontgomeryReduce(&dest.b, &ab2)           // = 2*a*b*R mod p

	return dest
}

// Set dest = lhs + rhs.
//
// Allowed to overlap lhs or rhs with dest.
//
// Returns dest to allow chaining operations.
func (dest *ExtensionFieldElement) Add(lhs, rhs *ExtensionFieldElement) *ExtensionFieldElement {
	fp751AddReduced(&dest.a, &lhs.a, &rhs.a)
	fp751AddReduced(&dest.b, &lhs.b, &rhs.b)

	return dest
}

// Set dest = lhs - rhs.
//
// Allowed to overlap lhs or rhs with dest.
//
// Returns dest to allow chaining operations.
func (dest *ExtensionFieldElement) Sub(lhs, rhs *ExtensionFieldElement) *ExtensionFieldElement {
	fp751SubReduced(&dest.a, &lhs.a, &rhs.a)
	fp751SubReduced(&dest.b, &lhs.b, &rhs.b)

	return dest
}

// Returns true if lhs = rhs.  Takes variable time.
func (lhs *ExtensionFieldElement) VartimeEq(rhs *ExtensionFieldElement) bool {
	return lhs.a.vartimeEq(rhs.a) && lhs.b.vartimeEq(rhs.b)
}

//------------------------------------------------------------------------------
// Prime Field
//------------------------------------------------------------------------------

// Represents an element of the prime field F_p.
type PrimeFieldElement struct {
	// This field element is in Montgomery form, so that the value `a` is
	// represented by `aR mod p`.
	a fp751Element
}

// Set dest to x.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) SetUint64(x uint64) *PrimeFieldElement {
	var xRR fp751X2
	dest.a = fp751Element{}                 // = 0
	dest.a[0] = x                           // = x
	fp751Mul(&xRR, &dest.a, &montgomeryRsq) // = x*R*R
	fp751MontgomeryReduce(&dest.a, &xRR)    // = x*R mod p

	return dest
}

// Set dest = lhs * rhs.
//
// Allowed to overlap lhs or rhs with dest.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) Mul(lhs, rhs *PrimeFieldElement) *PrimeFieldElement {
	a := &lhs.a // = a*R
	b := &rhs.a // = b*R

	var ab fp751X2
	fp751Mul(&ab, a, b)                 // = a*b*R*R
	fp751MontgomeryReduce(&dest.a, &ab) // = a*b*R mod p

	return dest
}

// Set dest = x^(2^k), for k >= 1, by repeated squarings.
//
// Allowed to overlap x with dest.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) Pow2k(x *PrimeFieldElement, k uint8) *PrimeFieldElement {
	dest.Square(x)
	for i := uint8(1); i < k; i++ {
		dest.Square(dest)
	}

	return dest
}

// Set dest = x^2
//
// Allowed to overlap x with dest.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) Square(x *PrimeFieldElement) *PrimeFieldElement {
	a := &x.a // = a*R
	b := &x.a // = b*R

	var ab fp751X2
	fp751Mul(&ab, a, b)                 // = a*b*R*R
	fp751MontgomeryReduce(&dest.a, &ab) // = a*b*R mod p

	return dest
}

// Set dest = lhs + rhs.
//
// Allowed to overlap lhs or rhs with dest.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) Add(lhs, rhs *PrimeFieldElement) *PrimeFieldElement {
	fp751AddReduced(&dest.a, &lhs.a, &rhs.a)

	return dest
}

// Set dest = lhs - rhs.
//
// Allowed to overlap lhs or rhs with dest.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) Sub(lhs, rhs *PrimeFieldElement) *PrimeFieldElement {
	fp751SubReduced(&dest.a, &lhs.a, &rhs.a)

	return dest
}

// Returns true if lhs = rhs.  Takes variable time.
func (lhs *PrimeFieldElement) VartimeEq(rhs *PrimeFieldElement) bool {
	return lhs.a.vartimeEq(rhs.a)
}

// Set dest = sqrt(x), if x is a square.  If x is nonsquare dest is undefined.
//
// Allowed to overlap x with dest.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) Sqrt(x *PrimeFieldElement) *PrimeFieldElement {
	tmp_x := *x // Copy x in case dest == x
	// Since x is assumed to be square, x = y^2
	dest.P34(x)            // dest = (y^2)^((p-3)/4) = y^((p-3)/2)
	dest.Mul(dest, &tmp_x) // dest = y^2 * y^((p-3)/2) = y^((p+1)/2)
	// Now dest^2 = y^(p+1) = y^2 = x, so dest = sqrt(x)

	return dest
}

// Set dest = 1/x.
//
// Allowed to overlap x with dest.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) Inv(x *PrimeFieldElement) *PrimeFieldElement {
	tmp_x := *x            // Copy x in case dest == x
	dest.Square(x)         // dest = x^2
	dest.P34(dest)         // dest = (x^2)^((p-3)/4) = x^((p-3)/2)
	dest.Square(dest)      // dest = x^(p-3)
	dest.Mul(dest, &tmp_x) // dest = x^(p-2)

	return dest
}

// Set dest = x^((p-3)/4)
//
// Allowed to overlap x with dest.
//
// Returns dest to allow chaining operations.
func (dest *PrimeFieldElement) P34(x *PrimeFieldElement) *PrimeFieldElement {
	// Sliding-window strategy computed with Sage, awk, sed, and tr.
	//
	// This performs sum(powStrategy) = 744 squarings and len(mulStrategy)
	// = 137 multiplications, in addition to 1 squaring and 15
	// multiplications to build a lookup table.
	//
	// In total this is 745 squarings, 152 multiplications.  Since squaring
	// is not implemented for the prime field, this is 897 multiplications
	// in total.
	powStrategy := [137]uint8{5, 7, 6, 2, 10, 4, 6, 9, 8, 5, 9, 4, 7, 5, 5, 4, 8, 3, 9, 5, 5, 4, 10, 4, 6, 6, 6, 5, 8, 9, 3, 4, 9, 4, 5, 6, 6, 2, 9, 4, 5, 5, 5, 7, 7, 9, 4, 6, 4, 8, 5, 8, 6, 6, 2, 9, 7, 4, 8, 8, 8, 4, 6, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 2}
	mulStrategy := [137]uint8{31, 23, 21, 1, 31, 7, 7, 7, 9, 9, 19, 15, 23, 23, 11, 7, 25, 5, 21, 17, 11, 5, 17, 7, 11, 9, 23, 9, 1, 19, 5, 3, 25, 15, 11, 29, 31, 1, 29, 11, 13, 9, 11, 27, 13, 19, 15, 31, 3, 29, 23, 31, 25, 11, 1, 21, 19, 15, 15, 21, 29, 13, 23, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 3}
	initialMul := uint8(27)

	// Build a lookup table of odd multiples of x.
	lookup := [16]PrimeFieldElement{}
	xx := &PrimeFieldElement{}
	xx.Square(x) // Set xx = x^2
	lookup[0] = *x
	for i := 1; i < 16; i++ {
		lookup[i].Mul(&lookup[i-1], xx)
	}
	// Now lookup = {x, x^3, x^5, ... }
	// so that lookup[i] = x^{2*i + 1}
	// so that lookup[k/2] = x^k, for odd k

	*dest = lookup[initialMul/2]
	for i := uint8(0); i < 137; i++ {
		dest.Pow2k(dest, powStrategy[i])
		dest.Mul(dest, &lookup[mulStrategy[i]/2])
	}

	return dest
}

//------------------------------------------------------------------------------
// Internals
//------------------------------------------------------------------------------

const fp751NumWords = 12

// (2^768) mod p.
// This can't be a constant because Go doesn't allow array constants, so try
// not to modify it.
var montgomeryR = fp751Element{149933, 0, 0, 0, 0, 9444048418595930112, 6136068611055053926, 7599709743867700432, 14455912356952952366, 5522737203492907350, 1222606818372667369, 49869481633250}

// (2^768)^2 mod p
// This can't be a constant because Go doesn't allow array constants, so try
// not to modify it.
var montgomeryRsq = fp751Element{2535603850726686808, 15780896088201250090, 6788776303855402382, 17585428585582356230, 5274503137951975249, 2266259624764636289, 11695651972693921304, 13072885652150159301, 4908312795585420432, 6229583484603254826, 488927695601805643, 72213483953973}

// Internal representation of an element of the base field F_p.
//
// This type is distinct from PrimeFieldElement in that no particular meaning
// is assigned to the representation -- it could represent an element in
// Montgomery form, or not.  Tracking the meaning of the field element is left
// to higher types.
type fp751Element [fp751NumWords]uint64

// Represents an intermediate product of two elements of the base field F_p.
type fp751X2 [2 * fp751NumWords]uint64

// Compute z = x + y (mod p).
//go:noescape
func fp751AddReduced(z, x, y *fp751Element)

// Compute z = x - y (mod p).
//go:noescape
func fp751SubReduced(z, x, y *fp751Element)

// Compute z = x + y, without reducing mod p.
//go:noescape
func fp751AddLazy(z, x, y *fp751Element)

// Compute z = x + y, without reducing mod p.
//go:noescape
func fp751X2AddLazy(z, x, y *fp751X2)

// Compute z = x - y, without reducing mod p.
//go:noescape
func fp751X2SubLazy(z, x, y *fp751X2)

// Compute z = x * y.
//go:noescape
func fp751Mul(z *fp751X2, x, y *fp751Element)

// Perform Montgomery reduction: set z = x R^{-1} (mod p).
// Destroys the input value.
//go:noescape
func fp751MontgomeryReduce(z *fp751Element, x *fp751X2)

// Reduce a field element in [0, 2*p) to one in [0,p).
//go:noescape
func fp751StrongReduce(x *fp751Element)

func (x fp751Element) vartimeEq(y fp751Element) bool {
	fp751StrongReduce(&x)
	fp751StrongReduce(&y)
	eq := true
	for i := 0; i < fp751NumWords; i++ {
		eq = (x[i] == y[i]) && eq
	}

	return eq
}
