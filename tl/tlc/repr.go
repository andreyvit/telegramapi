package tlc

import (
	"bytes"
	"github.com/andreyvit/telegramapi/tl/tlschema"
)

type Resolver interface {
	ResolveTypeExpr(expr tlschema.TypeExpr, context string) Repr
	FindType(name string) *tlschema.Type
	FindComb(name string) *tlschema.Comb
}

type GenericRepr interface {
	Specialize(typ tlschema.TypeExpr) Repr
	AppendSwitchCase(buf *bytes.Buffer, indent string)
	AppendGoDefs(buf *bytes.Buffer)
	Resolve(resolver Resolver)
	GoImports() []string
}

type Repr interface {
	Resolve(resolver Resolver)
	AppendReadStmt(buf *bytes.Buffer, indent, dst string)
	AppendWriteStmt(buf *bytes.Buffer, indent, src string)
	GoType() string
	GoImports() []string
}

type UnsupportedRepr struct {
	Name   string
	ErrMsg string
}

func (r *UnsupportedRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return r
}

func (r *UnsupportedRepr) Resolve(resolver Resolver) {
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

func (r *UnsupportedRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *UnsupportedRepr) GoType() string {
	return "interface{} /* " + r.Name + " - " + r.ErrMsg + " */"
}
func (r *UnsupportedRepr) GoImports() []string {
	return nil
}

type StringRepr struct {
}

func (r *StringRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *StringRepr) Resolve(resolver Resolver) {
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

func (r *StringRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *StringRepr) GoType() string {
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

func (r *BytesRepr) Resolve(resolver Resolver) {
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

func (r *BytesRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *BytesRepr) GoType() string {
	return "[]byte"
}
func (r *BytesRepr) GoImports() []string {
	return nil
}

type BigIntRepr struct {
}

func (r *BigIntRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *BigIntRepr) Resolve(resolver Resolver) {
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

func (r *BigIntRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *BigIntRepr) GoType() string {
	return "*big.Int"
}
func (r *BigIntRepr) GoImports() []string {
	return []string{"math/big"}
}

type UnixTimeRepr struct {
}

func (r *UnixTimeRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *UnixTimeRepr) Resolve(resolver Resolver) {
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

func (r *UnixTimeRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *UnixTimeRepr) GoType() string {
	return "time.Time"
}
func (r *UnixTimeRepr) GoImports() []string {
	return []string{"time"}
}

type IntRepr struct {
}

func (r *IntRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *IntRepr) Resolve(resolver Resolver) {
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

func (r *IntRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *IntRepr) GoType() string {
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

func (r *LongRepr) Resolve(resolver Resolver) {
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

func (r *LongRepr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *LongRepr) GoType() string {
	return "uint64"
}
func (r *LongRepr) GoImports() []string {
	return nil
}

type Int128Repr struct {
}

func (r *Int128Repr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *Int128Repr) Resolve(resolver Resolver) {
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

func (r *Int128Repr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *Int128Repr) GoType() string {
	return "[16]byte"
}
func (r *Int128Repr) GoImports() []string {
	return nil
}

type Int256Repr struct {
}

func (r *Int256Repr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyBare(r, typ)
}

func (r *Int256Repr) Resolve(resolver Resolver) {
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

func (r *Int256Repr) AppendGoDefs(buf *bytes.Buffer) {
}

func (r *Int256Repr) GoType() string {
	return "[32]byte"
}
func (r *Int256Repr) GoImports() []string {
	return nil
}

type ObjectRepr struct {
}

func (r *ObjectRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeOnlyNonBare(r, typ)
}

func (r *ObjectRepr) Resolve(resolver Resolver) {
}
func (r *ObjectRepr) AppendReadStmt(buf *bytes.Buffer, indent, dst string) {
	buf.WriteString(indent)
	buf.WriteString(dst)
	buf.WriteString(" = ReadBoxedObjectFrom(r)\n")
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
func (r *ObjectRepr) AppendGoDefs(buf *bytes.Buffer) {
}
func (r *ObjectRepr) GoType() string {
	return "tl.Object"
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
func (r *GenericVectorRepr) Resolve(resolver Resolver) {
	r.vectorComb = resolver.FindComb("vector")
	if r.vectorComb == nil {
		panic("vector constructor not found")
	}
}
func (r *GenericVectorRepr) AppendSwitchCase(buf *bytes.Buffer, indent string) {
}
func (r *GenericVectorRepr) AppendGoDefs(buf *bytes.Buffer) {
}
func (r *GenericVectorRepr) GoImports() []string {
	return nil
}

type VectorRepr struct {
	ItemType tlschema.TypeExpr

	ItemRepr Repr
}

func (r *VectorRepr) Resolve(resolver Resolver) {
	r.ItemRepr = resolver.ResolveTypeExpr(r.ItemType, "")
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

func (r *VectorRepr) GoType() string {
	return "[]" + r.ItemRepr.GoType()
}
func (r *VectorRepr) GoImports() []string {
	return append([]string{"errors"}, r.ItemRepr.GoImports()...)
}

type BoxedRepr struct {
	Comb     *tlschema.Comb
	ItemRepr Repr
}

func (r *BoxedRepr) Resolve(resolver Resolver) {
	// r.ItemRepr = resolver.ResolveTypeExpr(r.ItemType, "")
	r.ItemRepr.Resolve(resolver)
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
func (r *BoxedRepr) GoType() string {
	return r.ItemRepr.GoType()
}
func (r *BoxedRepr) GoImports() []string {
	return r.ItemRepr.GoImports()
}

type StructRepr struct {
	TLName string
	GoName string

	resolved bool
	Ctor     *tlschema.Comb
	ArgReprs []ArgRepr
}

type ArgRepr struct {
	// Arg    *tlschema.Arg
	TLName     string
	GoName     string
	TypeRepr   Repr
	TLTypeName string
}

func (r *StructRepr) Specialize(typ tlschema.TypeExpr) Repr {
	return specializeBare(r, r.Ctor, typ)
}

func (r *StructRepr) Resolve(resolver Resolver) {
	if r.resolved {
		return
	}
	r.resolved = true

	for _, arg := range r.Ctor.Args {
		ar := ArgRepr{
			// Arg:    &arg,
			TLName:     arg.Name,
			GoName:     tlschema.ToGoName(arg.Name),
			TypeRepr:   resolver.ResolveTypeExpr(arg.Type, r.TLName+":"+arg.Name),
			TLTypeName: arg.Type.String(),
		}
		ar.TypeRepr.Resolve(resolver)
		r.ArgReprs = append(r.ArgReprs, ar)
	}
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
	buf.WriteString("s := new(")
	buf.WriteString(r.GoName)
	buf.WriteString(")\n")

	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("s.ReadBareFrom(r)\n")

	buf.WriteString(indent)
	buf.WriteString(indent)
	buf.WriteString("return s\n")
}

func (r *StructRepr) AppendGoDefs(buf *bytes.Buffer) {
	buf.WriteString("\n")
	buf.WriteString("// ")
	buf.WriteString(r.GoName)
	buf.WriteString(" represents ")
	buf.WriteString(r.TLName)
	buf.WriteString(" from TL schema")
	buf.WriteString("\n")

	buf.WriteString("type ")
	buf.WriteString(r.GoName)
	buf.WriteString(" struct {\n")

	for _, ar := range r.ArgReprs {
		buf.WriteString("\t")
		buf.WriteString(ar.GoName)
		buf.WriteString(" ")
		buf.WriteString(ar.TypeRepr.GoType())
		buf.WriteString("  // ")
		buf.WriteString(ar.TLName)
		buf.WriteString(": ")
		buf.WriteString(ar.TLTypeName)
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")

	buf.WriteString("\n")
	buf.WriteString("func (s *")
	buf.WriteString(r.GoName)
	buf.WriteString(") Cmd() uint32 {\n")
	buf.WriteString("\treturn ")
	buf.WriteString(IDConstName(r.Ctor))
	buf.WriteString(";\n")
	buf.WriteString("}\n")

	buf.WriteString("\n")
	buf.WriteString("func (s *")
	buf.WriteString(r.GoName)
	buf.WriteString(") ReadBareFrom(r *tl.Reader) {\n")
	for _, ar := range r.ArgReprs {
		ar.TypeRepr.AppendReadStmt(buf, "\t", "s."+ar.GoName)
	}
	buf.WriteString("}\n")

	buf.WriteString("\n")
	buf.WriteString("func (s *")
	buf.WriteString(r.GoName)
	buf.WriteString(") WriteBareTo(w *tl.Writer) {\n")
	for _, ar := range r.ArgReprs {
		ar.TypeRepr.AppendWriteStmt(buf, "\t", "s."+ar.GoName)
	}
	buf.WriteString("}\n")
}

func (r *StructRepr) GoType() string {
	return "*" + r.GoName
}
func (r *StructRepr) GoImports() []string {
	var result []string
	for _, ar := range r.ArgReprs {
		result = append(result, ar.TypeRepr.GoImports()...)
	}
	return result
}

// func (r *StructRepr) GoDef(buf *bytes.Buffer) {
// 	// return "*" + r.GoName
// }
