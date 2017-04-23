package tlc

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/andreyvit/telegramapi/tl/tlschema"
)

type CodeGenOptions struct {
	SkipUtil     bool
	SkipSwitch   bool
	SkipComments bool
}

type Resolver interface {
	ResolveTypeExpr(expr tlschema.TypeExpr, context string) Repr
	AddContributor(c Contributor) Contributor
	FindComb(name string) *tlschema.Comb
}

type Contributor interface {
	InternalTypeID() string
	Resolve(resolver Resolver) error

	AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions)
	GoImports() []string
}

type GenericRepr interface {
	Contributor
	Specialize(typ tlschema.TypeExpr) Repr
	AppendSwitchCase(buf *bytes.Buffer, indent string)
}

type Repr interface {
	Contributor
	AppendReadStmt(buf *bytes.Buffer, indent, dst string)
	AppendWriteStmt(buf *bytes.Buffer, indent, src string)
	GoType() string
}

type UnsupportedRepr struct {
	Name   string
	ErrMsg string
}

func (r *UnsupportedRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return r
}
func (r *UnsupportedRepr) Resolve(resolver Resolver) error {
	return nil
}
func (r *UnsupportedRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString("// TODO: read ")
	buf.WriteString(dst)
	buf.WriteString("\n")
}
func (r *UnsupportedRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("// TODO: write ")
	buf.WriteString(src)
	buf.WriteString("\n")
}
func (r *UnsupportedRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}
func (r *UnsupportedRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *UnsupportedRepr) GoType() string {
	return "interface{} /* " + r.Name + " - " + r.ErrMsg + " */"
}
func (r *UnsupportedRepr) InternalTypeID() string {
	return "Unsupported-" + r.Name
}
func (r *UnsupportedRepr) GoImports() []string {
	return nil
}

type StringRepr struct {
}

func (r *StringRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *StringRepr) Resolve(resolver Resolver) error {
	return nil
}

func (r *StringRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("r.ReadString()")
	buf.WriteString("\n")
}

func (r *StringRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteString(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *StringRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *StringRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}

func (r *StringRepr) GoType() string {
	return "string"
}
func (r *StringRepr) InternalTypeID() string {
	return "string"
}
func (r *StringRepr) GoImports() []string {
	return nil
}

type BytesRepr struct {
}

func (r *BytesRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *BytesRepr) Resolve(resolver Resolver) error {
	return nil
}

func (r *BytesRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("r.ReadBlob()")
	buf.WriteString("\n")
}

func (r *BytesRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteBlob(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *BytesRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *BytesRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}

func (r *BytesRepr) GoType() string {
	return "[]byte"
}
func (r *BytesRepr) InternalTypeID() string {
	return "bytes"
}
func (r *BytesRepr) GoImports() []string {
	return nil
}

type BigIntRepr struct {
}

func (r *BigIntRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *BigIntRepr) Resolve(resolver Resolver) error {
	return nil
}

func (r *BigIntRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("r.ReadBigInt()")
	buf.WriteString("\n")
}

func (r *BigIntRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteBigInt(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *BigIntRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *BigIntRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}

func (r *BigIntRepr) GoType() string {
	return "*big.Int"
}
func (r *BigIntRepr) InternalTypeID() string {
	return "bigint"
}
func (r *BigIntRepr) GoImports() []string {
	return []string{"math/big"}
}

type UnixTimeRepr struct {
}

func (r *UnixTimeRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *UnixTimeRepr) Resolve(resolver Resolver) error {
	return nil
}

func (r *UnixTimeRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("r.ReadTimeSec32()")
	buf.WriteString("\n")
}

func (r *UnixTimeRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteTimeSec32(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *UnixTimeRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *UnixTimeRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}

func (r *UnixTimeRepr) GoType() string {
	return "time.Time"
}
func (r *UnixTimeRepr) InternalTypeID() string {
	return "unixtime"
}
func (r *UnixTimeRepr) GoImports() []string {
	return []string{"time"}
}

type NatRepr struct {
}

func (r *NatRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}
func (r *NatRepr) Resolve(resolver Resolver) error {
	return nil
}
func (r *NatRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("uint(r.ReadUint32())")
	buf.WriteString("\n")
}
func (r *NatRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteUint32(uint32(")
	buf.WriteString(src)
	buf.WriteString("))\n")
}
func (r *NatRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}
func (r *NatRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *NatRepr) GoType() string {
	return "uint"
}
func (r *NatRepr) InternalTypeID() string {
	return "nat"
}
func (r *NatRepr) GoImports() []string {
	return nil
}

type IntRepr struct {
}

func (r *IntRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}
func (r *IntRepr) Resolve(resolver Resolver) error {
	return nil
}
func (r *IntRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("r.ReadInt()")
	buf.WriteString("\n")
}
func (r *IntRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteInt(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}
func (r *IntRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}
func (r *IntRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *IntRepr) GoType() string {
	return "int"
}
func (r *IntRepr) InternalTypeID() string {
	return "int"
}
func (r *IntRepr) GoImports() []string {
	return nil
}

type LongRepr struct {
}

func (r *LongRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *LongRepr) Resolve(resolver Resolver) error {
	return nil
}

func (r *LongRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("r.ReadUint64()")
	buf.WriteString("\n")
}

func (r *LongRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteUint64(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}

func (r *LongRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *LongRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}

func (r *LongRepr) GoType() string {
	return "uint64"
}
func (r *LongRepr) InternalTypeID() string {
	return "long"
}
func (r *LongRepr) GoImports() []string {
	return nil
}

type Int128Repr struct {
}

func (r *Int128Repr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *Int128Repr) Resolve(resolver Resolver) error {
	return nil
}

func (r *Int128Repr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString("r.ReadUint128(")
	buf.WriteString(dst)
	buf.WriteString("[:])\n")
}

func (r *Int128Repr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteUint128(")
	buf.WriteString(src)
	buf.WriteString("[:])\n")
}

func (r *Int128Repr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *Int128Repr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}

func (r *Int128Repr) GoType() string {
	return "[16]byte"
}
func (r *Int128Repr) InternalTypeID() string {
	return "int128"
}
func (r *Int128Repr) GoImports() []string {
	return nil
}

type Int256Repr struct {
}

func (r *Int256Repr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *Int256Repr) Resolve(resolver Resolver) error {
	return nil
}

func (r *Int256Repr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString("r.ReadFull(")
	buf.WriteString(dst)
	buf.WriteString("[:])\n")
}

func (r *Int256Repr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.Write(")
	buf.WriteString(src)
	buf.WriteString("[:])\n")
}

func (r *Int256Repr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *Int256Repr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}

func (r *Int256Repr) GoType() string {
	return "[32]byte"
}
func (r *Int256Repr) InternalTypeID() string {
	return "int256"
}
func (r *Int256Repr) GoImports() []string {
	return nil
}

type DoubleRepr struct {
}

func (r *DoubleRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}
func (r *DoubleRepr) Resolve(resolver Resolver) error {
	return nil
}
func (r *DoubleRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("r.ReadFloat64()")
	buf.WriteString("\n")
}
func (r *DoubleRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteFloat64(")
	buf.WriteString(src)
	buf.WriteString(")\n")
}
func (r *DoubleRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *DoubleRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}

func (r *DoubleRepr) GoType() string {
	return "float64"
}
func (r *DoubleRepr) InternalTypeID() string {
	return "double"
}
func (r *DoubleRepr) GoImports() []string {
	return nil
}

type BoolRepr struct {
	trueComb  *tlschema.Comb
	falseComb *tlschema.Comb
}

func (r *BoolRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyNonBare(r, typ)
}
func (r *BoolRepr) Resolve(resolver Resolver) error {
	r.trueComb = resolver.FindComb("boolTrue")
	if r.trueComb == nil {
		return errors.New("'true' constructor not found")
	}
	r.falseComb = resolver.FindComb("boolFalse")
	if r.falseComb == nil {
		return errors.New("'true' constructor not found")
	}
	return nil
}
func (r *BoolRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString("r.ExpectCmd(")
	buf.WriteString(IDConstName(r.trueComb))
	buf.WriteString(", ")
	buf.WriteString(IDConstName(r.falseComb))
	buf.WriteString(")\n")

	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ")
	buf.WriteString("(r.ReadCmd() == ")
	buf.WriteString(IDConstName(r.trueComb))
	buf.WriteString(")\n")
}
func (r *BoolRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("if ")
	buf.WriteString(src)
	buf.WriteString("{\n")

	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("w.WriteCmd(")
	buf.WriteString(IDConstName(r.trueComb))
	buf.WriteString(")\n")

	buf.WriteString(indent)
	buf.WriteString("} else {\n")

	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("w.WriteCmd(")
	buf.WriteString(IDConstName(r.falseComb))
	buf.WriteString(")\n")

	buf.WriteString(indent)
	buf.WriteString("}\n")
}
func (r *BoolRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *BoolRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *BoolRepr) GoType() string {
	return "bool"
}
func (r *BoolRepr) InternalTypeID() string {
	return "bool"
}
func (r *BoolRepr) GoImports() []string {
	return nil
}

type TrueRepr struct {
	trueComb *tlschema.Comb
}

func (r *TrueRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeBare(r, r.trueComb, typ)
}
func (r *TrueRepr) Resolve(resolver Resolver) error {
	r.trueComb = resolver.FindComb("true")
	if r.trueComb == nil {
		return errors.New("'true' constructor not found")
	}
	return nil
}
func (r *TrueRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = true\n")
}
func (r *TrueRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
}
func (r *TrueRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}
func (r *TrueRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *TrueRepr) GoType() string {
	return "bool"
}
func (r *TrueRepr) InternalTypeID() string {
	return "true"
}
func (r *TrueRepr) GoImports() []string {
	return nil
}

type ObjectRepr struct {
}

func (r *ObjectRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyNonBare(r, typ)
}

func (r *ObjectRepr) Resolve(resolver Resolver) error {
	return nil
}
func (r *ObjectRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = Schema.ReadBoxedObjectFrom(r)\n")
}
func (r *ObjectRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	// buf.WriteString(indent)
	// buf.WriteString("if ")
	// buf.WriteString(src)
	// buf.WriteString(" != nil {\b")

	buf.WriteString(indent)
	buf.WriteString("w.WriteCmd(")
	buf.WriteString(src)
	buf.WriteString(".Cmd())\n")

	buf.WriteString(indent)
	buf.WriteString(src)
	buf.WriteString(".WriteBareTo(w)\n")

	// buf.WriteString(indent)
	// buf.WriteString("}\n")
}
func (r *ObjectRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}
func (r *ObjectRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *ObjectRepr) GoType() string {
	return "tl.Object"
}
func (r *ObjectRepr) InternalTypeID() string {
	return "Object"
}
func (r *ObjectRepr) GoImports() []string {
	return nil
}

type GenericVectorRepr struct {
	vectorComb *tlschema.Comb
}

func (r *GenericVectorRepr) Specialize(typ tlschema.TypeExpr) Repr {
	if len(typ.GenericArgs) != 1 {
		return nil
	}
	vec := &VectorRepr{ItemType: typ.GenericArgs[0]}
	// return vec
	if r.vectorComb == nil {
		panic("vector not resolved")
	}
	return specializeBare(vec, r.vectorComb, typ)
}
func (r *GenericVectorRepr) Resolve(resolver Resolver) error {
	r.vectorComb = resolver.FindComb("vector")
	if r.vectorComb == nil {
		return errors.New("vector constructor not found")
	}
	return nil
}
func (r *GenericVectorRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}
func (r *GenericVectorRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *GenericVectorRepr) InternalTypeID() string {
	return "Vector"
}
func (r *GenericVectorRepr) GoImports() []string {
	return nil
}

type VectorRepr struct {
	ItemType tlschema.TypeExpr

	ItemRepr Repr
}

func (r *VectorRepr) Resolve(resolver Resolver) error {
	r.ItemRepr = resolver.ResolveTypeExpr(r.ItemType, "")
	return nil
}
func (r *VectorRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = make([]")
	buf.WriteString(r.ItemRepr.GoType())
	buf.WriteString(", r.ReadInt())")
	buf.WriteString("\n")

	buf.WriteString(indent)
	buf.WriteString("for i := 0; i < len(")
	buf.WriteString(dst)
	buf.WriteString("); i++ {\n")
	r.ItemRepr.AppendReadStmt(buf, indent+indent, dst+"[i]")
	buf.WriteString(indent)
	buf.WriteString("}\n")
}
func (r *VectorRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteInt(len(")
	buf.WriteString(src)
	buf.WriteString("))\n")

	buf.WriteString(indent)
	buf.WriteString("for i := 0; i < len(")
	buf.WriteString(src)
	buf.WriteString("); i++ {\n")
	r.ItemRepr.AppendWriteStmt(buf, indent+indent, src+"[i]")
	buf.WriteString(indent)
	buf.WriteString("}\n")
}
func (r *VectorRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *VectorRepr) GoType() string {
	return "[]" + r.ItemRepr.GoType()
}
func (r *VectorRepr) InternalTypeID() string {
	return "Vector<" + r.ItemType.String() + ">"
}
func (r *VectorRepr) GoImports() []string {
	return []string{"errors"}
}

type BoxedRepr struct {
	Comb     *tlschema.Comb
	ItemRepr Repr
}

func (r *BoxedRepr) Resolve(resolver Resolver) error {
	// r.ItemRepr = resolver.ResolveTypeExpr(r.ItemType, "")
	r.ItemRepr = resolver.AddContributor(r.ItemRepr).(Repr)
	return nil
}
func (r *BoxedRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString("if cmd := r.ReadCmd(); cmd != ")
	buf.WriteString(IDConstName(r.Comb))
	buf.WriteString("{\n")
	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("r.Fail(errors.New(\"expected: ")
	buf.WriteString(r.Comb.CombName.Full())
	buf.WriteString("\"))\n")
	buf.WriteString(indent)
	buf.WriteString("}\n")
	r.ItemRepr.AppendReadStmt(buf, indent, dst)
}

func (r *BoxedRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteCmd(")
	buf.WriteString(IDConstName(r.Comb))
	buf.WriteString(")\n")
	r.ItemRepr.AppendWriteStmt(buf, indent, src)
}
func (r *BoxedRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
}
func (r *BoxedRepr) GoType() string {
	return r.ItemRepr.GoType()
}
func (r *BoxedRepr) InternalTypeID() string {
	return "Box<" + r.ItemRepr.InternalTypeID() + ">"
}
func (r *BoxedRepr) GoImports() []string {
	return nil
}

type StructRepr struct {
	TLName string
	GoName string
	Ctor   *tlschema.Comb

	GoMarkerFuncName string

	ArgReprs []*ArgRepr
}

type ArgCondType int

const (
	NoCond ArgCondType = iota
	PureFlag
	FieldWithFlag
)

type ArgRepr struct {
	Arg        tlschema.Arg
	TLName     string
	GoName     string
	TypeRepr   Repr
	TLTypeName string

	CondType ArgCondType
	CondArg  *ArgRepr
	CondBit  int

	CondGetterGoName string
	CondSetterGoName string
}

func (ar *ArgRepr) IsCond() bool {
	return ar.CondType != NoCond
}

func (ar *ArgRepr) HasField() bool {
	return ar.CondType != PureFlag
}

func (r *StructRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeBare(r, r.Ctor, typ)
}

func (r *StructRepr) findArg(tlName string) *ArgRepr {
	for _, ar := range r.ArgReprs {
		if ar.TLName == tlName {
			return ar
		}
	}
	return nil
}

func (r *StructRepr) findConflictingGoName(name string) bool {
	for _, ar := range r.ArgReprs {
		if ar.HasField() && (ar.GoName == name) {
			return true
		}
		if ar.IsCond() {
			if ar.CondGetterGoName == name {
				return true
			}
			if ar.CondSetterGoName == name {
				return true
			}
		}
	}
	return false
}

func (r *StructRepr) Resolve(resolver Resolver) error {
	for _, arg := range r.Ctor.Args {
		ar := &ArgRepr{
			Arg:        arg,
			TLName:     arg.Name,
			GoName:     tlschema.ToGoName(arg.Name),
			TypeRepr:   resolver.ResolveTypeExpr(arg.Type, r.TLName+":"+arg.Name),
			TLTypeName: arg.Type.String(),
		}
		if arg.CondArgName != "" {
			ca := r.findArg(arg.CondArgName)
			if ca == nil {
				return fmt.Errorf("Cannot find argument specified as conditional: %q", arg.CondArgName)
			}
			ar.CondArg = ca
			if _, ok := ar.TypeRepr.(*TrueRepr); ok {
				ar.CondType = PureFlag
				ar.CondGetterGoName = ar.GoName
				ar.CondSetterGoName = "Set" + ar.GoName
			} else {
				ar.CondType = FieldWithFlag
				ar.CondGetterGoName = "Has" + ar.GoName
				ar.CondSetterGoName = "SetHas" + ar.GoName
				if r.findConflictingGoName(ar.CondGetterGoName) {
					ar.CondGetterGoName = "Has" + ar.GoName + "Field"
					ar.CondSetterGoName = "SetHas" + ar.GoName + "Field"
				}
			}
			ar.CondBit = arg.CondBit
		}
		r.ArgReprs = append(r.ArgReprs, ar)
	}

	return nil
}

func (r *StructRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = new(")
	buf.WriteString(r.GoName)
	buf.WriteString(")\n")

	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(".ReadBareFrom(r)")
	buf.WriteString("\n")
}

func (r *StructRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString(src)
	buf.WriteString(".WriteBareTo(w)\n")
}

func (r *StructRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
	buf.WriteString(indent)
	buf.WriteString("case ")
	buf.WriteString(IDConstName(r.Ctor))
	buf.WriteString(":\n")

	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("return new(")
	buf.WriteString(r.GoName)
	buf.WriteString(")\n")
}

func (r *StructRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
	buf.WriteString("\n")
	if !options.SkipComments {
		buf.WriteString("// ")
		buf.WriteString(r.GoName)
		buf.WriteString(" represents ")
		buf.WriteString(r.Ctor.String())
		buf.WriteString(" from ")
		buf.WriteString(r.Ctor.Origin)
		buf.WriteString("\n")
	}

	buf.WriteString("type ")
	buf.WriteString(r.GoName)
	buf.WriteString(" struct {\n")

	for _, ar := range r.ArgReprs {
		if !ar.HasField() {
			continue
		}
		buf.WriteString("\t")
		buf.WriteString(ar.GoName)
		buf.WriteString(" ")
		buf.WriteString(ar.TypeRepr.GoType())
		if !options.SkipComments {
			buf.WriteString("  // ")
			buf.WriteString(ar.Arg.String())
		}
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")

	if r.GoMarkerFuncName != "" {
		buf.WriteString("\n")
		buf.WriteString("func (o *")
		buf.WriteString(r.GoName)
		buf.WriteString(") ")
		buf.WriteString(r.GoMarkerFuncName)
		buf.WriteString("() {}\n")
	}

	buf.WriteString("\n")
	buf.WriteString("func (o *")
	buf.WriteString(r.GoName)
	buf.WriteString(") Cmd() uint32 {\n")
	buf.WriteString("\treturn ")
	buf.WriteString(IDConstName(r.Ctor))
	buf.WriteString(";\n")
	buf.WriteString("}\n")

	buf.WriteString("\n")
	buf.WriteString("func (o *")
	buf.WriteString(r.GoName)
	buf.WriteString(") ReadBareFrom(r *tl.Reader) {\n")
	for _, ar := range r.ArgReprs {
		if ar.TLTypeName == "true" {
			continue
		}
		subindent := "\t"
		if ar.IsCond() {
			buf.WriteString(fmt.Sprintf("\tif (o.%s&(1<<%d)) != 0 {\n", ar.CondArg.GoName, ar.Arg.CondBit))
			subindent = "\t\t"
		}
		ar.TypeRepr.AppendReadStmt(buf, subindent, "o."+ar.GoName)
		if ar.IsCond() {
			buf.WriteString("\t}\n")
		}
	}
	buf.WriteString("}\n")

	buf.WriteString("\n")
	buf.WriteString("func (o *")
	buf.WriteString(r.GoName)
	buf.WriteString(") WriteBareTo(w *tl.Writer) {\n")
	for _, ar := range r.ArgReprs {
		if ar.TLTypeName == "true" {
			continue
		}
		subindent := "\t"
		if ar.IsCond() {
			buf.WriteString(fmt.Sprintf("\tif (o.%s&(1<<%d)) != 0 {\n", ar.CondArg.GoName, ar.Arg.CondBit))
			subindent = "\t\t"
		}
		ar.TypeRepr.AppendWriteStmt(buf, subindent, "o."+ar.GoName)
		if ar.IsCond() {
			buf.WriteString("\t}\n")
		}
	}
	buf.WriteString("}\n")

	for _, ar := range r.ArgReprs {
		if !ar.IsCond() {
			continue
		}
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("func (o *%s) %s() bool {\n", r.GoName, ar.CondGetterGoName))
		buf.WriteString(fmt.Sprintf("return (o.%s & (1<<%d)) != 0\n}\n", ar.CondArg.GoName, ar.CondBit))
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("func (o *%s) %s(v bool) {\n", r.GoName, ar.CondSetterGoName))
		buf.WriteString(fmt.Sprintf("if v { o.%s |= (1<<%d) } else { o.%s &= ^%s(1<<%d) }\n}\n", ar.CondArg.GoName, ar.CondBit, ar.CondArg.GoName, ar.CondArg.TypeRepr.GoType(), ar.CondBit))
	}

	if !options.SkipUtil {
		buf.WriteString("\n")
		buf.WriteString("func (o *")
		buf.WriteString(r.GoName)
		buf.WriteString(") String() string {\n")
		buf.WriteString("\treturn tl.Pretty(o)\n")
		buf.WriteString("}\n")
	}
}

func (r *StructRepr) GoType() string {
	return "*" + r.GoName
}
func (r *StructRepr) InternalTypeID() string {
	return r.TLName
}
func (r *StructRepr) GoImports() []string {
	return nil
}

// func (r *StructRepr) GoDef(buf *bytes.Buffer) {
// 	// return "*" + r.GoName
// }

type MultiCtorRepr struct {
	TLName           string
	GoName           string
	GoMarkerFuncName string
	Structs          []*StructRepr
}

func (r *MultiCtorRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyNonBare(r, typ)
}

func (r *MultiCtorRepr) Resolve(resolver Resolver) error {
	for _, struc := range r.Structs {
		resolver.AddContributor(struc)
	}
	return nil
}

func (r *MultiCtorRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = Schema.ReadLimitedBoxedObjectFrom(r")
	for _, struc := range r.Structs {
		buf.WriteString(", ")
		buf.WriteString(IDConstName(struc.Ctor))
	}
	buf.WriteString(").(")
	buf.WriteString(r.GoName)
	buf.WriteString(")\n")
}

func (r *MultiCtorRepr) AppendWriteStmt(buf *bytes.Buffer, indent, src string) {
	buf.WriteString(indent)
	buf.WriteString("w.WriteCmd(")
	buf.WriteString(src)
	buf.WriteString(".Cmd())\n")
	buf.WriteString(indent)
	buf.WriteString(src)
	buf.WriteString(".WriteBareTo(w)\n")
}

func (r *MultiCtorRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}

func (r *MultiCtorRepr) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
	buf.WriteString("\n")
	if !options.SkipComments {
		buf.WriteString("// ")
		buf.WriteString(r.GoName)
		buf.WriteString(" represents ")
		buf.WriteString(r.TLName)
		buf.WriteString(" from ")
		buf.WriteString(r.Structs[0].Ctor.Origin)
		buf.WriteString("\n")
	}
	buf.WriteString("type ")
	buf.WriteString(r.GoName)
	buf.WriteString(" interface {\n")
	buf.WriteString("\t")
	buf.WriteString(r.GoMarkerFuncName)
	buf.WriteString("()\n")
	buf.WriteString("\tCmd() uint32\n")
	buf.WriteString("\tReadBareFrom(r *tl.Reader)\n")
	buf.WriteString("\tWriteBareTo(w *tl.Writer)\n")
	buf.WriteString("}\n")
}

func (r *MultiCtorRepr) GoType() string {
	return r.GoName
}
func (r *MultiCtorRepr) InternalTypeID() string {
	return r.TLName
}
func (r *MultiCtorRepr) GoImports() []string {
	return nil
}
